<h1 align="center">mcdiskfree</h1>

<p align="center">
  <b>See what is eating your disk.</b><br>
  A free, open-source disk space analyzer for macOS, Windows and Linux.
</p>

<p align="center">
  <a href="https://github.com/mormitech/mcdiskfree/actions/workflows/build.yml"><img alt="build" src="https://github.com/mormitech/mcdiskfree/actions/workflows/build.yml/badge.svg"></a>
  <a href="LICENSE"><img alt="MIT license" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
  <img alt="platforms" src="https://img.shields.io/badge/platforms-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg">
  <img alt="status" src="https://img.shields.io/badge/status-early%20preview-orange.svg">
</p>

---

## Our goal

> **A disk analyzer that is free, forever, on every platform, and installed the way you install everything else.**

```
brew install --cask mormitech/tap/mcdiskfree      # macOS            (planned)
brew install mormitech/tap/mcdiskfree             # Linux            (planned)
winget install mormitech.mcdiskfree               # Windows          (planned)
```

Good disk analyzers exist, but the best ones are paid and tied to one operating system. mcdiskfree is an attempt to
build the same experience (a clear, interactive picture of where your space went) as a small native app that costs
nothing, has no accounts, no ads, no telemetry and no upsell, and is installable with the package manager you already use.

## What it does

mcdiskfree scans a folder or a whole disk and draws it as an interactive **sunburst chart**: every ring is a folder
level, every slice is sized by the space it takes on disk.

- **Click to drill in.** Click a slice to open the folder, the centre (or **↑ Up**, or right-click) to go back, **⌂ Root** to jump to the top.
- **Whole-disk view.** The start screen lists your disks with their usage; scan one and the free space shows up as a dashed slice.
- **Free space always visible.** Size and free space of the disk you are looking at, in the header and in the centre of the chart.
- **A tree next to the chart.** Open folders with the triangle without leaving your place in the chart; every row shows its share.
- **A basket, and a safe delete.** Collect files and folders (hover a row → **+ Basket**, or shift-click a slice), then **Move to Trash**.
  Things only ever go to the Trash / Recycle Bin, never a permanent delete, and only from inside the folder you scanned.
- **Instant reopen.** Scans are cached; opening a disk you scanned before shows it at once ("scanned 2 h ago"). Rescan everything, or just one folder.
- **Picks up where you left off.** Restarting restores the disk and folder you were looking at.
- **Native window, one small binary.** Not a browser tab, not a 200 MB Electron app.

## Install

mcdiskfree is an **early preview** (version 0.5), so the package-manager releases are not published yet; that is the
next milestone (see the roadmap). Until then, build it from source, it takes a minute:

```
git clone https://github.com/mormitech/mcdiskfree
cd mcdiskfree/frontend && npm install && cd ..
wails build            # the app ends up in build/bin
```

You need [Go](https://go.dev), Node and the [Wails](https://wails.io) CLI
(`go install github.com/wailsapp/wails/v2/cmd/wails@latest`). On Linux also `libgtk-3-dev libwebkit2gtk-4.1-dev`, and
build with `-tags webkit2_41`.

### Platform status

| | macOS | Windows | Linux |
|---|---|---|---|
| Builds in CI | ✅ | ✅ | ✅ |
| Used and tried by hand | ✅ (Apple silicon) | ⏳ not yet | ⏳ not yet |
| Move to Trash | Finder (so "Put Back" works) | Recycle Bin (untested) | `gio trash` or `trash-put` (untested) |
| Package manager | Homebrew (planned) | winget (planned) | Homebrew (planned) |

Releases are not code-signed yet, so the OS will warn on the first launch (macOS: **Open Anyway** in System Settings →
Privacy & Security; Windows: **More info → Run anyway**).

## Privacy

mcdiskfree makes no network connections and sends nothing anywhere. The only thing it writes is its cache: a
compressed copy of your last scans (file and folder names and sizes) in your user cache folder
(`~/Library/Caches/mcdiskfree`, `%LocalAppData%\mcdiskfree`, `~/.cache/mcdiskfree`). Delete that folder any time.

## How it works

- **Scanner (Go).** Walks the tree in parallel, counts the space files take **on disk** (allocated blocks, not nominal
  size), skips symlinks and does not cross into other mounted volumes.
- **Chart and UI (vanilla JS + d3).** A webview inside a native window (Wails); `d3-hierarchy` and `d3-shape` draw the sunburst.
- **Cache.** Each full scan is saved as a gzip-compressed file, standard library only, no database. The biggest 100
  entries per folder are kept, the rest becomes "smaller items". A folder's modification time is deliberately **not**
  used to skip work: it does not change when a file inside grows, so a cached size could silently be wrong. Instead the
  age is always shown and one folder can be rescanned and merged in.
- **OS specifics** (volume list, free space, Trash) live in [`internal/disk`](internal/disk), one small file per platform.

## Roadmap

Next, in this order:

1. **Ship it.** Signed releases, Homebrew tap, winget package, first real tests on Windows and Linux. *(This is the goal.)*
2. **Honest numbers.** Detect folders it may not read (macOS Full Disk Access) and say so; show purgeable space and
   snapshots so the chart adds up to the disk; handle APFS clones, hard links and iCloud placeholders.
3. **Live scan.** Watch the chart grow while it scans.
4. **Quality of life.** App icon, search, keyboard navigation, drag a folder onto the window, file-type colours,
   Quick Look, short explanations for system folders, units and theme settings.
5. **Real change detection** with file-system events, for instant incremental updates.

Have an idea or found a bug? [Open an issue](https://github.com/mormitech/mcdiskfree/issues).

## Develop

```
cd frontend && npm install && cd ..
wails dev          # live-reloading window
go test ./...
```

| Path | What |
|---|---|
| `scan.go` | parallel scanner, tree helpers |
| `cache.go`, `session.go` | saved scans and the last session |
| `app.go` | the methods the UI calls |
| `internal/disk` | volumes, free space, Trash, per OS (cross-compiles on its own) |
| `frontend/` | the UI |
| `.github/workflows/build.yml` | builds and tests on macOS, Windows and Linux |

The version lives in one place, `wails.json`.

## License

[MIT](LICENSE): use it, change it, ship it, forever.
