package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"katdown/internal/config"
	"katdown/internal/editor"
	"katdown/internal/render"
	"katdown/internal/watch"
)

// docChanged tells the frontend to fetch the document again.
const docChanged = "doc:changed"

var markdownFilter = []application.FileFilter{
	{DisplayName: "Markdown", Pattern: "*.md;*.markdown;*.mdown;*.mkd;*.txt"},
	{DisplayName: "All files", Pattern: "*"},
}

// App is bound to the frontend; its exported methods are callable from JS.
type App struct {
	wails  *application.App
	window *application.WebviewWindow

	// rawHTML passes raw HTML in documents through to the page (-html).
	rawHTML bool

	mu      sync.Mutex
	path    string
	watcher *watch.Watcher
}

// Document is everything the frontend needs to display the current file.
type Document struct {
	Path      string  `json:"path"`
	Name      string  `json:"name"`
	HTML      string  `json:"html"`
	CSS       string  `json:"css"`
	CustomCSS bool    `json:"customCss"`
	Theme     string  `json:"theme"` // dark, light or system
	CodeCSS   CodeCSS `json:"codeCss"`
	Error     string  `json:"error"`
	// ConfigError describes a problem with config.toml; defaults were used
	// for the affected settings.
	ConfigError string `json:"configError"`
}

// CodeCSS is the syntax-highlighting stylesheet for each theme.
type CodeCSS struct {
	Dark  string `json:"dark"`
	Light string `json:"light"`
}

func NewApp(path string, rawHTML bool) *App {
	return &App{path: path, rawHTML: rawHTML}
}

// domReady runs once the page has loaded: it shows the file from the command
// line, or asks for one and quits if none is chosen.
func (a *App) domReady() {
	a.mu.Lock()
	p := a.path
	a.mu.Unlock()

	if p != "" {
		a.setFile(p)
		return
	}
	p, err := a.wails.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title:          "Open markdown file",
		CanChooseFiles: true,
		Filters:        markdownFilter,
		Window:         a.window,
	}).PromptForSingleSelection()
	if err != nil || p == "" {
		a.wails.Quit()
		return
	}
	a.setFile(p)
}

// ServiceShutdown is called by Wails when the app quits.
func (a *App) ServiceShutdown() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.watcher != nil {
		a.watcher.Close()
	}
	return nil
}

// setFile makes p the displayed file, watches it, and tells the frontend.
func (a *App) setFile(p string) {
	a.mu.Lock()
	a.path = p
	if a.watcher != nil {
		a.watcher.Close()
		a.watcher = nil
	}
	w, err := watch.File(p, func() { a.wails.Event.Emit(docChanged) })
	if err != nil {
		log.Printf("watch %s: %v", p, err)
	}
	a.watcher = w
	a.mu.Unlock()

	a.window.SetTitle("katdown — " + filepath.Base(p))
	a.wails.Event.Emit(docChanged)
}

func (a *App) currentPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.path
}

// GetDocument reads and renders the current file with the current settings
// and stylesheet. Both are re-read every time, so a reload picks up edits.
func (a *App) GetDocument() Document {
	settings, err := config.LoadSettings()
	css, custom := config.LoadCSS()
	dark, light := settings.CodeCSS()
	doc := Document{
		// The settings' variables come last so they apply to a custom style.css too.
		CSS:       css + "\n" + settings.CSSVars(),
		CustomCSS: custom,
		Theme:     settings.Theme,
		CodeCSS:   CodeCSS{Dark: dark, Light: light},
	}
	if err != nil {
		doc.ConfigError = err.Error()
	}

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
	html, err := render.Render(src, settings.HighlightPatterns(), a.rawHTML)
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
	p, err := a.wails.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title:          "Open markdown file",
		Directory:      dir,
		CanChooseFiles: true,
		Filters:        markdownFilter,
		Window:         a.window,
	}).PromptForSingleSelection()
	if err != nil || p == "" {
		return
	}
	a.setFile(p)
}

// ResolveRelative returns the file a link such as "other.md#intro" points to,
// relative to the current file's directory. The fragment is dropped; the
// frontend scrolls to it once the file is shown.
func (a *App) ResolveRelative(href string) (string, error) {
	cur := a.currentPath()
	if cur == "" {
		return "", fmt.Errorf("no file open")
	}
	target, err := url.PathUnescape(strings.SplitN(href, "#", 2)[0])
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(cur), filepath.FromSlash(target))
	}
	if _, err := os.Stat(target); err != nil {
		return "", err
	}
	return target, nil
}

// OpenPath displays the file at p, an absolute path such as one returned by
// ResolveRelative or an earlier Document.Path.
func (a *App) OpenPath(p string) error {
	if _, err := os.Stat(p); err != nil {
		return err
	}
	a.setFile(p)
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
	if err := a.wails.Browser.OpenURL(u); err != nil {
		log.Printf("katdown: opening %s: %v", u, err)
	}
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

// EditFile opens the current file in the user's editor; see editor.Command.
// The watcher reloads the view whenever the editor saves.
func (a *App) EditFile() error {
	p := a.currentPath()
	if p == "" {
		return fmt.Errorf("no file open")
	}
	settings, _ := config.LoadSettings()
	cmd, err := editor.Command(settings.Editor.Command, p, editor.Env{
		Getenv:   os.Getenv,
		LookPath: exec.LookPath,
		GOOS:     goruntime.GOOS,
		WSL:      isWSL(),
	})
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	// An editor that can't run (say, a terminal editor with no terminal)
	// exits at once; report that rather than failing silently.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			if msg := strings.TrimSpace(stderr.String()); msg != "" {
				return fmt.Errorf("%v: %s", err, msg)
			}
			return err
		}
	case <-time.After(editorGrace):
	}
	return nil
}

// editorGrace is how long EditFile waits to see whether the editor fails.
const editorGrace = time.Second

// Quit closes the app.
func (a *App) Quit() {
	a.wails.Quit()
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
