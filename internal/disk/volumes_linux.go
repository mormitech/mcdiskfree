package disk

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var realFS = map[string]bool{
	"ext2": true, "ext3": true, "ext4": true, "xfs": true, "btrfs": true, "f2fs": true, "zfs": true, "jfs": true,
	"ntfs": true, "ntfs3": true, "vfat": true, "exfat": true, "fuseblk": true, "hfsplus": true,
}

// Volumes lists the mounted disks backed by a real file system, once per device.
func Volumes() []Volume {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return rootOnly()
	}
	defer f.Close()
	var out []Volume
	seen := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 || !realFS[fields[2]] || seen[fields[0]] {
			continue
		}
		dev, mnt := fields[0], strings.ReplaceAll(fields[1], `\040`, " ")
		if strings.HasPrefix(mnt, "/snap") || strings.HasPrefix(mnt, "/var/lib/docker") {
			continue
		}
		u, err := UsageOf(mnt)
		if err != nil || u.Total == 0 {
			continue
		}
		seen[dev] = true
		name := filepath.Base(mnt)
		if mnt == "/" {
			name = "Root"
		}
		out = append(out, Volume{Name: name, Path: mnt, Total: u.Total, Free: u.Free})
	}
	if len(out) == 0 {
		return rootOnly()
	}
	return out
}

func rootOnly() []Volume {
	u, err := UsageOf("/")
	if err != nil {
		return nil
	}
	return []Volume{{Name: "Root", Path: "/", Total: u.Total, Free: u.Free}}
}
