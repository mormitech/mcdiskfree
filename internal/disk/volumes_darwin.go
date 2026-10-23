package disk

import (
	"os"
	"path/filepath"
)

// Volumes lists the internal data volume and everything mounted under /Volumes.
//
// The startup disk is shown as its Data volume: that is where the user's files live, and scanning "/" would either skip
// them (they are firmlinked in from another device) or count them twice.
func Volumes() []Volume {
	var out []Volume
	add := func(name, path string) {
		if u, err := UsageOf(path); err == nil && u.Total > 0 {
			out = append(out, Volume{Name: name, Path: path, Total: u.Total, Free: u.Free})
		}
	}
	add("Macintosh HD", "/System/Volumes/Data")
	entries, _ := os.ReadDir("/Volumes")
	for _, e := range entries {
		p := filepath.Join("/Volumes", e.Name())
		if info, err := os.Lstat(p); err != nil || info.Mode()&os.ModeSymlink != 0 {
			continue // "/Volumes/Macintosh HD" is a symlink to "/"
		}
		add(e.Name(), p)
	}
	return out
}
