package disk

import (
	"errors"
	"fmt"
	"os/exec"
)

// Trash moves path to the freedesktop.org trash with `gio trash` (or trash-cli's `trash-put`).
func Trash(path string) error {
	for _, c := range [][]string{{"gio", "trash", "--"}, {"trash-put", "--"}} {
		if _, err := exec.LookPath(c[0]); err != nil {
			continue
		}
		if out, err := exec.Command(c[0], append(c[1:], path)...).CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %s", path, out)
		}
		return nil
	}
	return errors.New("no trash tool found: install gio (glib) or trash-cli")
}
