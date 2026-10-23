package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanSizesAndFind(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "a", "b"), 0o755)
	os.WriteFile(filepath.Join(dir, "a", "b", "big"), make([]byte, 100000), 0o644)
	os.WriteFile(filepath.Join(dir, "small"), make([]byte, 10), 0o644)
	os.Symlink(filepath.Join(dir, "a"), filepath.Join(dir, "link"))

	s := &scanner{ctx: context.Background(), sem: make(chan struct{}, 4)}
	root := s.walk(dir, "root")

	if root.Items != 2 {
		t.Fatalf("items = %d, want 2 (symlink must not be followed)", root.Items)
	}
	if root.Children[0].Name != "a" {
		t.Fatalf("biggest child = %q, want a", root.Children[0].Name)
	}
	b := find(root, filepath.Join(dir, "a", "b"))
	if b == nil || b.Size < 100000 {
		t.Fatalf("find(a/b) = %+v", b)
	}
	if find(root, filepath.Join(dir, "..", "x")) != nil {
		t.Fatal("find must not leave the tree")
	}
	p := prune(root, 1)
	if len(p.Children[0].Children) != 0 {
		t.Fatal("prune depth 1 must not keep grandchildren")
	}
}

func TestRemoveFromTree(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "a", "b"), 0o755)
	os.WriteFile(filepath.Join(dir, "a", "b", "big"), make([]byte, 100000), 0o644)
	os.WriteFile(filepath.Join(dir, "a", "keep"), make([]byte, 5000), 0o644)

	s := &scanner{ctx: context.Background(), sem: make(chan struct{}, 4)}
	root := s.walk(dir, "root")
	before := root.Size
	b := find(root, filepath.Join(dir, "a", "b"))
	gone := b.Size

	if !removeFromTree(root, b.Path) {
		t.Fatal("remove failed")
	}
	if root.Size != before-gone || find(root, filepath.Join(dir, "a")).Size != before-gone {
		t.Fatalf("sizes not reduced: root %d, want %d", root.Size, before-gone)
	}
	if root.Items != 1 || find(root, b.Path) != nil {
		t.Fatal("node still in tree")
	}
	if removeFromTree(root, dir) {
		t.Fatal("the root itself must not be removable")
	}
}
