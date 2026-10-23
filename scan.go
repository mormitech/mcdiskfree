package main

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// Node is one file or folder in the scanned tree. Size is the space it occupies on disk, in bytes.
type Node struct {
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Size     int64   `json:"size"`
	IsDir    bool    `json:"isDir"`
	Items    int64   `json:"items"`
	Children []*Node `json:"children,omitempty"`
}

type scanner struct {
	ctx     context.Context
	sem     chan struct{}
	device  uint64 // file system of the scan root; other mounts below it are skipped (0 = no check)
	scanned atomic.Int64
}

func (s *scanner) walk(path, name string) *Node {
	n := &Node{Name: name, Path: path, IsDir: true}
	// The semaphore is held only while reading the directory, never while waiting for children, so deep trees cannot deadlock.
	s.sem <- struct{}{}
	entries, err := os.ReadDir(path)
	var infos []os.FileInfo
	if err == nil {
		for _, e := range entries {
			if info, err := e.Info(); err == nil {
				infos = append(infos, info)
			}
		}
	}
	<-s.sem

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, info := range infos {
		if s.ctx.Err() != nil {
			break
		}
		child := filepath.Join(path, info.Name())
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			continue
		case info.IsDir():
			if s.device != 0 && deviceOf(info) != s.device {
				continue // another volume mounted inside this one
			}
			wg.Add(1)
			go func(child, name string) {
				defer wg.Done()
				c := s.walk(child, name)
				mu.Lock()
				n.Children = append(n.Children, c)
				mu.Unlock()
			}(child, info.Name())
		case info.Mode().IsRegular():
			s.scanned.Add(1)
			mu.Lock()
			n.Children = append(n.Children, &Node{Name: info.Name(), Path: child, Size: diskSize(info), Items: 1})
			mu.Unlock()
		}
	}
	wg.Wait()

	for _, c := range n.Children {
		n.Size += c.Size
		n.Items += c.Items
	}
	sort.Slice(n.Children, func(i, j int) bool { return n.Children[i].Size > n.Children[j].Size })
	return n
}

// find returns the node at path inside the tree rooted at root, or nil.
func find(root *Node, path string) *Node {
	rel, err := filepath.Rel(root.Path, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil
	}
	cur := root
	if rel == "." {
		return cur
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		var next *Node
		for _, c := range cur.Children {
			if c.Name == part {
				next = c
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

const (
	viewDepth    = 4  // rings shown in the chart
	viewChildren = 24 // biggest entries kept per folder; the rest is folded into one "smaller items" entry
)

// prune copies n down to depth levels, folding the small entries of each folder into one.
func prune(n *Node, depth int) *Node {
	out := &Node{Name: n.Name, Path: n.Path, Size: n.Size, IsDir: n.IsDir, Items: n.Items}
	if depth == 0 || len(n.Children) == 0 {
		return out
	}
	var rest *Node
	for i, c := range n.Children {
		if i < viewChildren {
			out.Children = append(out.Children, prune(c, depth-1))
			continue
		}
		if rest == nil {
			rest = &Node{IsDir: false}
		}
		rest.Size += c.Size
		rest.Items += c.Items
	}
	if rest != nil && rest.Size > 0 {
		rest.Name = "smaller items"
		out.Children = append(out.Children, rest)
	}
	return out
}

// locate returns the folders from root down to the parent of path, and the index of path among that parent's children.
func locate(root *Node, path string) (chain []*Node, idx int, ok bool) {
	rel, err := filepath.Rel(root.Path, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, 0, false
	}
	parts := strings.Split(rel, string(filepath.Separator))
	chain = []*Node{root}
	for _, part := range parts[:len(parts)-1] {
		var next *Node
		for _, c := range chain[len(chain)-1].Children {
			if c.Name == part {
				next = c
				break
			}
		}
		if next == nil {
			return nil, 0, false
		}
		chain = append(chain, next)
	}
	for i, c := range chain[len(chain)-1].Children {
		if c.Name == parts[len(parts)-1] {
			return chain, i, true
		}
	}
	return nil, 0, false
}

// removeFromTree drops the node at path and takes its size off every folder above it.
func removeFromTree(root *Node, path string) bool {
	chain, i, ok := locate(root, path)
	if !ok {
		return false
	}
	parent := chain[len(chain)-1]
	gone := parent.Children[i]
	parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
	for _, a := range chain {
		a.Size -= gone.Size
		a.Items -= gone.Items
	}
	return true
}

// replaceNode swaps the node at path for fresh (a new scan of the same folder) and corrects the sizes above it.
func replaceNode(root *Node, path string, fresh *Node) bool {
	chain, i, ok := locate(root, path)
	if !ok {
		return false
	}
	parent := chain[len(chain)-1]
	old := parent.Children[i]
	parent.Children[i] = fresh
	for k := len(chain) - 1; k >= 0; k-- {
		chain[k].Size += fresh.Size - old.Size
		chain[k].Items += fresh.Items - old.Items
		kids := chain[k].Children
		sort.SliceStable(kids, func(x, y int) bool { return kids[x].Size > kids[y].Size })
	}
	return true
}
