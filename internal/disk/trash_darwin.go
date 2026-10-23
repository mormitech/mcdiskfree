package disk

import (
	"fmt"
	"os/exec"
)

// Trash moves path to the Trash through Finder, so "Put Back" works. The first call makes macOS ask whether the app may
// control Finder. The path travels as an argument, never inside the script text.
func Trash(path string) error {
	script := []string{
		"-e", "on run argv",
		"-e", `tell application "Finder" to delete (POSIX file (item 1 of argv) as alias)`,
		"-e", "end run",
	}
	if out, err := exec.Command("osascript", append(script, path)...).CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", path, out)
	}
	return nil
}
