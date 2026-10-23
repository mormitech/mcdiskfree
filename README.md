# mcdiskfree

A free, cross-platform disk space analyzer. It scans a folder or a whole disk and shows what takes up the space as an
interactive sunburst chart, in a native window on macOS, Windows and Linux.

- Click a folder ring to drill in; click the centre, right-click or press **↑ Up** to go back.
- Shows the free space of the disk, and the free slice when you scan a whole disk.
- **Basket:** collect files and folders (hover a row → **+ Basket**, or shift-click a ring), then **Move to Trash**.
  Items only ever go to the Trash / Recycle Bin, never a permanent delete.

## Build

Needs [Go](https://go.dev), Node and the [Wails](https://wails.io) CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).
On Linux also `libgtk-3-dev libwebkit2gtk-4.1-dev`, and build with `-tags webkit2_41`.

```
cd frontend && npm install && cd ..
wails dev      # live-reloading window
wails build    # binary or .app in build/bin
go test ./...
```

## Layout

- `scan.go`: parallel scanner, sizes are space on disk, symlinks and other mounts are not followed.
- `app.go`: the methods the UI calls (scan, view, volumes, free space, reveal, trash).
- `internal/disk`: the per-OS parts (volume list, free space, Trash), cross-compilable on its own.
- `frontend/`: vanilla JS + `d3-hierarchy` / `d3-shape`.

The macOS Trash goes through Finder (so "Put Back" works); the first time, macOS asks whether the app may control Finder.
Releases are not code-signed yet, so the OS warns on first launch.
