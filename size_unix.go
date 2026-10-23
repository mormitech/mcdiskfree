//go:build !windows

package main

import (
	"os"
	"syscall"
)

// diskSize is the space the file occupies on disk (allocated blocks), which is what a full disk cares about.
func diskSize(info os.FileInfo) int64 {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return int64(st.Blocks) * 512
	}
	return info.Size()
}

// deviceOf identifies the file system a file lives on, so the scan can stay on one volume.
func deviceOf(info os.FileInfo) uint64 {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return uint64(st.Dev)
	}
	return 0
}
