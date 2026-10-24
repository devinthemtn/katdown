package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"katdown/internal/config"
	"katdown/internal/render"
	"katdown/internal/watch"
)

// docChanged tells the frontend to fetch the document again.
const docChanged = "doc:changed"

var markdownFilter = []runtime.FileFilter{
	{DisplayName: "Markdown", Pattern: "*.md;*.markdown;*.mdown;*.mkd;*.txt"},
	{DisplayName: "All files", Pattern: "*"},
}

// App is bound to the frontend; its exported methods are callable from JS.
type App struct {
	ctx context.Context

	mu      sync.Mutex
	path    string
	watcher *watch.Watcher
}

// Document is everything the frontend needs to display the current file.
type Document struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	HTML      string `json:"html"`
	CSS       string `json:"css"`
	CustomCSS bool   `json:"customCss"`
	Error     string `json:"error"`
}

func NewApp(path string) *App {
	return &App{path: path}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) domReady(ctx context.Context) {
	a.mu.Lock()
	p := a.path
	a.mu.Unlock()

	if p != "" {
		a.setFile(p)
		return
	}
	p, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title:   "Open markdown file",
		Filters: markdownFilter,
	})
	if err != nil || p == "" {
		runtime.Quit(ctx)
		return
	}
	a.setFile(p)
}

func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.watcher != nil {
		a.watcher.Close()
	}
}

// setFile makes p the displayed file, watches it, and tells the frontend.
func (a *App) setFile(p string) {
	a.mu.Lock()
	a.path = p
	if a.watcher != nil {
		a.watcher.Close()
		a.watcher = nil
	}
	w, err := watch.File(p, func() { runtime.EventsEmit(a.ctx, docChanged) })
	if err != nil {
		log.Printf("watch %s: %v", p, err)
	}
	a.watcher = w
	a.mu.Unlock()

	runtime.WindowSetTitle(a.ctx, "katdown — "+filepath.Base(p))
	runtime.EventsEmit(a.ctx, docChanged)
}

func (a *App) currentPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.path
}

// GetDocument reads and renders the current file with the active stylesheet.
func (a *App) GetDocument() Document {
	css, custom := config.LoadCSS()
	doc := Document{CSS: css, CustomCSS: custom}

	p := a.currentPath()
	if p == "" {
		return doc
	}
	doc.Path, doc.Name = p, filepath.Base(p)

	src, err := os.ReadFile(p)
	if err != nil {
		doc.Error = err.Error()
		return doc
	}
	html, err := render.Render(src)
	if err != nil {
		doc.Error = err.Error()
		return doc
	}
	doc.HTML = html
	return doc
}

// OpenFile shows the file dialog and displays the chosen file.
func (a *App) OpenFile() {
	dir := ""
	if p := a.currentPath(); p != "" {
		dir = filepath.Dir(p)
	}
	p, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Open markdown file",
		DefaultDirectory: dir,
		Filters:          markdownFilter,
	})
	if err != nil || p == "" {
		return
	}
	a.setFile(p)
}

// OpenRelative opens a link target relative to the current file's directory,
// such as another .md file linked from this one.
func (a *App) OpenRelative(href string) error {
	cur := a.currentPath()
	if cur == "" {
		return fmt.Errorf("no file open")
	}
	target, err := url.PathUnescape(strings.SplitN(href, "#", 2)[0])
	if err != nil {
		return err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(cur), filepath.FromSlash(target))
	}
	if _, err := os.Stat(target); err != nil {
		return err
	}
	a.setFile(target)
	return nil
}

// OpenExternal opens a URL in the system browser. Under WSL there is usually
// no xdg-open, so the URL is handed to Windows to open in its default browser.
func (a *App) OpenExternal(u string) {
	if isWSL() {
		if err := openWindows(u); err == nil {
			return
		} else {
			log.Printf("katdown: opening %s via Windows: %v", u, err)
		}
	}
	runtime.BrowserOpenURL(a.ctx, u)
}

func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	b, err := os.ReadFile("/proc/sys/kernel/osrelease")
	return err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft")
}

// openWindows prefers wslview (wslu) and falls back to explorer.exe, which
// opens a URL in the Windows default browser.
func openWindows(u string) error {
	name := "wslview"
	if _, err := exec.LookPath(name); err != nil {
		name = "explorer.exe"
	}
	cmd := exec.Command(name, u)
	if err := cmd.Start(); err != nil {
		return err
	}
	// explorer.exe exits non-zero even on success, so only reap the process.
	go cmd.Wait()
	return nil
}

// Quit closes the app.
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

// localFiles serves files next to the current document (e.g. relative images)
// for requests the embedded frontend assets don't match. It is unexported so
// Wails doesn't bind it to the frontend.
func (a *App) localFiles(w http.ResponseWriter, r *http.Request) {
	cur := a.currentPath()
	if cur == "" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	root := filepath.Dir(cur)
	rel := filepath.FromSlash(strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/"))
	full := filepath.Join(root, rel)
	if back, err := filepath.Rel(root, full); err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, full)
}
