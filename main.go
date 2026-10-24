package main

import (
	"embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"katdown/internal/config"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	makeConfig := flag.Bool("makeconfig", false, "create the katdown config folder with the default style.css, then exit")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: katdown [-makeconfig] [file.md]\n\n")
		flag.PrintDefaults()
		if p, err := config.CSSPath(); err == nil {
			fmt.Fprintf(flag.CommandLine.Output(), "\nCustom stylesheet: %s\n", p)
		}
	}
	flag.Parse()

	if *makeConfig {
		path, created, err := config.MakeConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, "katdown:", err)
			os.Exit(1)
		}
		if created {
			fmt.Println("Wrote", path)
		} else {
			fmt.Println(path, "already exists, left untouched")
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

	app := NewApp(path)
	err := wails.Run(&options.App{
		Title:  "katdown",
		Width:  960,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: http.HandlerFunc(app.localFiles),
		},
		BackgroundColour: &options.RGBA{R: 13, G: 16, B: 23, A: 255},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
