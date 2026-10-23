//go:build darwin || linux

package disk

import "syscall"

// UsageOf reports the volume that holds path.
func UsageOf(path string) (Usage, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return Usage{}, err
	}
	bs := uint64(st.Bsize)
	return Usage{Total: st.Blocks * bs, Free: st.Bavail * bs}, nil
}
