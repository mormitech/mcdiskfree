package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session is what was on screen when the app last ran: the scanned disk or folder and the folder shown inside it.
type Session struct {
	Root string `json:"root"`
	Path string `json:"path"`
}

func sessionFile() string {
	dir := cacheDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "session.json")
}

// SaveSession remembers what is shown, so the next start can open it again. An empty root means the list of disks.
func (a *App) SaveSession(root, path string) {
	if a.mock != nil {
		return // demo mode leaves the real session alone
	}
	file := sessionFile()
	if file == "" {
		return
	}
	data, err := json.Marshal(Session{Root: root, Path: path})
	if err != nil || os.MkdirAll(filepath.Dir(file), 0o755) != nil {
		return
	}
	tmp := file + ".tmp"
	if os.WriteFile(tmp, data, 0o644) == nil {
		os.Rename(tmp, file)
	}
}

// LoadSession returns the last session, or an empty one when there is none or its disk or folder is gone.
func (a *App) LoadSession() Session {
	if a.mock != nil {
		return Session{}
	}
	file := sessionFile()
	if file == "" {
		return Session{}
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return Session{}
	}
	var s Session
	if json.Unmarshal(data, &s) != nil {
		return Session{}
	}
	if _, err := os.Stat(s.Root); err != nil {
		return Session{}
	}
	return s
}
