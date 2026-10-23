//go:build !darwin && !linux && !windows

package disk

import "errors"

func Trash(path string) error {
	return errors.New("moving to the trash is not supported on this platform")
}
