package main

import (
	"context"

	wr "github.com/wailsapp/wails/v2/pkg/runtime"
)

// fitSize shrinks the wanted window size so it fits a screen of the given logical size. The screen size does not
// exclude the taskbar or dock, so only 90% of it is used.
func fitSize(wantW, wantH, screenW, screenH int) (int, int) {
	if screenW <= 0 || screenH <= 0 {
		return wantW, wantH
	}
	return min(wantW, screenW*9/10), min(wantH, screenH*9/10)
}

// fitWindow makes the main window fit the screen it is on and centres it, so the title bar stays reachable when the
// desktop scaling is high. Sizes are logical pixels, the same unit WindowSetSize takes.
func fitWindow(ctx context.Context, wantW, wantH int) {
	screens, err := wr.ScreenGetAll(ctx)
	if err != nil {
		return
	}
	for _, s := range screens {
		if !s.IsCurrent {
			continue
		}
		w, h := fitSize(wantW, wantH, s.Size.Width, s.Size.Height)
		if w != wantW || h != wantH {
			wr.WindowSetSize(ctx, w, h)
			wr.WindowCenter(ctx)
		}
		return
	}
}
