//go:build !darwin && !linux && !windows

package disk

import "errors"

func UsageOf(path string) (Usage, error) { return Usage{}, errors.New("unsupported platform") }

func Volumes() []Volume { return nil }
