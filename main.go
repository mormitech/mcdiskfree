package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

// wails.json is the single source of the version: it ends up in the macOS Info.plist and the Windows file properties,
// and is read from here for the window title and the UI.
//
//go:embed wails.json
var projectJSON []byte

func version() string {
	var p struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if json.Unmarshal(projectJSON, &p) != nil || p.Info.ProductVersion == "" {
		return "dev"
	}
	return p.Info.ProductVersion
}

// demoFlag reads --mock <platform> (or --mock=<platform>, or --demo): the platform of the fictional disks to show
// instead of the real ones. It returns "" when demo mode is off. Other arguments, such as the ones macOS adds when
// it launches an app, are ignored.
func demoFlag(args []string) string {
	var platform string
	fs := flag.NewFlagSet("mcdiskfree", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&platform, "mock", "", "show fictional disks of this platform (windows, linux or macos) instead of the real ones")
	fs.StringVar(&platform, "demo", "", "same as --mock")
	fs.Parse(args)
	return platform
}

func main() {
	var mock *mockFS
	if platform := demoFlag(os.Args[1:]); platform != "" {
		var err error
		if mock, err = loadMock(platform); err != nil {
			fmt.Fprintln(os.Stderr, "mcdiskfree:", err)
			os.Exit(2)
		}
	}
	app := NewApp(mock)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "mcdiskfree v" + version(),
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		// Without Mac options Wails leaves the green zoom (maximise) button disabled.
		Mac:       &mac.Options{},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
