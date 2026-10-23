package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	wr "github.com/wailsapp/wails/v2/pkg/runtime"
	"mcdiskfree/internal/disk"
)

// App is bound to the frontend: every exported method can be called from JavaScript.
type App struct {
	ctx    context.Context
	mu     sync.Mutex
	root   *Node
	cancel context.CancelFunc
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// Home is the folder the app offers to scan first.
func (a *App) Home() string {
	h, _ := os.UserHomeDir()
	return h
}

// ChooseFolder opens the OS folder picker. It returns "" when the user cancels.
func (a *App) ChooseFolder() (string, error) {
	return wr.OpenDirectoryDialog(a.ctx, wr.OpenDialogOptions{Title: "Folder to scan", DefaultDirectory: a.Home()})
}

// Scan walks path and returns the top of the tree. Progress is sent as "progress" events (files seen so far).
func (a *App) Scan(path string) (*Node, error) {
	a.Cancel()
	ctx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mu.Unlock()
	defer cancel()

	s := &scanner{ctx: ctx, sem: make(chan struct{}, 32)}
	if info, err := os.Lstat(path); err == nil {
		s.device = deviceOf(info)
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		t := time.NewTicker(150 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				wr.EventsEmit(a.ctx, "progress", s.scanned.Load())
			}
		}
	}()

	root := s.walk(path, filepath.Base(path))
	if ctx.Err() != nil {
		return nil, errors.New("scan cancelled")
	}
	a.mu.Lock()
	a.root = root
	a.mu.Unlock()
	return prune(root, viewDepth), nil
}

// Cancel stops a running scan.
func (a *App) Cancel() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

// View returns the part of the scanned tree at path (from the scan already done, so it is instant).
func (a *App) View(path string) (*Node, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.root == nil {
		return nil, errors.New("nothing scanned yet")
	}
	n := find(a.root, path)
	if n == nil {
		return nil, errors.New("not in the scanned tree: " + path)
	}
	return prune(n, viewDepth), nil
}

// Reveal shows path in the OS file manager.
func (a *App) Reveal(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "-R", path).Start()
	case "windows":
		return exec.Command("explorer", "/select,", path).Start()
	default:
		return exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}

// Volumes lists the disks that can be scanned as a whole, with their free space.
func (a *App) Volumes() []disk.Volume { return disk.Volumes() }

// FreeSpace reports the size and free space of the volume that holds path.
func (a *App) FreeSpace(path string) (disk.Usage, error) { return disk.UsageOf(path) }

// TrashResult says which paths were moved to the Trash and why the others were not.
type TrashResult struct {
	Trashed []string `json:"trashed"`
	Failed  []string `json:"failed"`
}

// Trash moves paths from the scanned tree to the Trash / Recycle Bin (never a permanent delete) and takes them out of
// the tree. Only paths inside the scanned folder are accepted, never the scanned folder itself.
func (a *App) Trash(paths []string) TrashResult {
	var res TrashResult
	for _, p := range paths {
		a.mu.Lock()
		ok := a.root != nil && p != a.root.Path && find(a.root, p) != nil
		a.mu.Unlock()
		if !ok {
			res.Failed = append(res.Failed, p+": not inside the scanned folder")
			continue
		}
		if err := disk.Trash(p); err != nil {
			res.Failed = append(res.Failed, err.Error())
			continue
		}
		a.mu.Lock()
		removeFromTree(a.root, p)
		a.mu.Unlock()
		res.Trashed = append(res.Trashed, p)
	}
	return res
}
