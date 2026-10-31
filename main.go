package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"

	"katdown/internal/config"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	makeConfig := flag.Bool("makeconfig", false, "create config.toml with the default settings, then exit")
	makeStyle := flag.Bool("makestyle", false, "create style.css with the default stylesheet, then exit")
	rawHTML := flag.Bool("html", false, "don't sanitize raw HTML in the document (allows scripts and styles); only for files you trust")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: katdown [-html] [-makeconfig] [-makestyle] [file.md]\n\n")
		flag.PrintDefaults()
		if p, err := config.SettingsPath(); err == nil {
			fmt.Fprintf(flag.CommandLine.Output(), "\nSettings:          %s\n", p)
		}
		if p, err := config.CSSPath(); err == nil {
			fmt.Fprintf(flag.CommandLine.Output(), "Custom stylesheet: %s\n", p)
		}
	}
	flag.Parse()

	if *makeConfig || *makeStyle {
		if *makeConfig {
			writeConfigFile(config.MakeSettings)
		}
		if *makeStyle {
			writeConfigFile(config.MakeStyle)
		}
		return
	}

	var path string
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}
	if flag.NArg() == 1 {
		abs, err := filepath.Abs(flag.Arg(0))
		if err == nil {
			_, err = os.Stat(abs)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "katdown:", err)
			os.Exit(1)
		}
		path = abs
	}

	// First run: write the default config.toml so there is a file to edit.
	// MakeSettings never overwrites, and a read-only config folder just means
	// the defaults are used.
	if _, _, err := config.MakeSettings(); err != nil {
		log.Printf("katdown: creating settings file: %v", err)
	}
	// Problems with the file are shown in the window once it opens.
	settings, _ := config.LoadSettings()
	bg := &options.RGBA{R: 13, G: 16, B: 23, A: 255}
	if settings.Theme == "light" {
		bg = &options.RGBA{R: 251, G: 251, B: 250, A: 255}
	}

	app := NewApp(path, *rawHTML)
	err := wails.Run(&options.App{
		Title:  "katdown",
		Width:  settings.Window.Width,
		Height: settings.Window.Height,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: http.HandlerFunc(app.localFiles),
		},
		BackgroundColour: bg,
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		Linux: &linux.Options{
			Icon:        icon,
			ProgramName: "katdown",
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// writeConfigFile runs one of the config.Make* functions and reports the result.
func writeConfigFile(write func() (string, bool, error)) {
	path, created, err := write()
	if err != nil {
		fmt.Fprintln(os.Stderr, "katdown:", err)
		os.Exit(1)
	}
	if created {
		fmt.Println("Wrote", path)
	} else {
		fmt.Println(path, "already exists, left untouched")
	}
}
