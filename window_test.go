package main

import "testing"

func TestFitSize(t *testing.T) {
	cases := []struct{ sw, sh, w, h int }{
		{1920, 1080, 1280, 800}, // roomy screen: unchanged
		{1093, 614, 983, 552},   // 1366x768 at 125%
		{1280, 720, 1152, 648},  // 720p
		{0, 0, 1280, 800},       // unknown screen: unchanged
	}
	for _, c := range cases {
		if w, h := fitSize(1280, 800, c.sw, c.sh); w != c.w || h != c.h {
			t.Errorf("fitSize on %dx%d = %dx%d, want %dx%d", c.sw, c.sh, w, h, c.w, c.h)
		}
	}
}
