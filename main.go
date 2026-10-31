package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

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
	bg := application.NewRGB(13, 16, 23)
	if settings.Theme == "light" {
		bg = application.NewRGB(251, 251, 250)
	}

	app := NewApp(path, *rawHTML)
	app.wails = application.New(application.Options{
		Name:     "katdown",
		Icon:     icon,
		Services: []application.Service{application.NewService(app)},
		Assets: application.AssetOptions{
			Handler: withFallback(application.AssetFileServerFS(assets), http.HandlerFunc(app.localFiles)),
		},
		Linux: application.LinuxOptions{
			ProgramName: "katdown",
		},
	})
	app.window = app.wails.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "katdown",
		Width:            settings.Window.Width,
		Height:           settings.Window.Height,
		BackgroundColour: bg,
		URL:              "/",
	})
	app.window.OnWindowEvent(events.Common.WindowRuntimeReady, func(*application.WindowEvent) {
		app.domReady()
	})
	if err := app.wails.Run(); err != nil {
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

// withFallback serves requests with h, and hands those h answers with 404 to
// fallback instead, like Wails v2's AssetServer.Handler did.
func withFallback(h, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fw := &fallbackWriter{ResponseWriter: w}
		h.ServeHTTP(fw, r)
		if fw.notFound {
			fallback.ServeHTTP(w, r)
		}
	})
}

// fallbackWriter passes a response through unless its status is 404, in
// which case it drops the status and body so the fallback can answer.
type fallbackWriter struct {
	http.ResponseWriter
	wrote    bool
	notFound bool
}

func (w *fallbackWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.wrote = true
	if code == http.StatusNotFound {
		w.notFound = true
		clear(w.Header())
		return
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *fallbackWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	if w.notFound {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}
