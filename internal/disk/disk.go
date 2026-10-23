// Package disk holds the operating-system specific parts: the list of volumes, the free space of a path and moving
// files to the Trash / Recycle Bin. It has no UI dependencies, so it can be cross-compiled and tested on its own.
package disk

// Volume is a disk (or partition) the user can scan as a whole.
type Volume struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

// Usage is the size and free space of the volume a path lives on.
type Usage struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}
