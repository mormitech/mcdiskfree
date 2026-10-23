package disk

import (
	"golang.org/x/sys/windows"
)

// UsageOf reports the volume that holds path.
func UsageOf(path string) (Usage, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return Usage{}, err
	}
	var avail, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(p, &avail, &total, &free); err != nil {
		return Usage{}, err
	}
	return Usage{Total: total, Free: avail}, nil
}

// Volumes lists the fixed and removable drives by letter.
func Volumes() []Volume {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil
	}
	var out []Volume
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		root := string(rune('A'+i)) + `:\`
		p, _ := windows.UTF16PtrFromString(root)
		if t := windows.GetDriveType(p); t != windows.DRIVE_FIXED && t != windows.DRIVE_REMOVABLE {
			continue
		}
		if u, err := UsageOf(root); err == nil && u.Total > 0 {
			out = append(out, Volume{Name: root, Path: root, Total: u.Total, Free: u.Free})
		}
	}
	return out
}
