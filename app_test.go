package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestListFiles(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"notes.md", "notes.txt", "next.markdown", "image.png", ".hidden.md", "sub/deep.md"} {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	home := filepath.Join(dir, "sub")

	tests := []struct {
		prefix string
		want   []string
	}{
		{"", []string{"next.markdown", "notes.md", "notes.txt", "sub/"}},
		{"no", []string{"notes.md", "notes.txt"}},
		{".", []string{".hidden.md"}},
		{"sub/", []string{"sub/deep.md"}},
		{"sub/d", []string{"sub/deep.md"}},
		{"~/", []string{"~/deep.md"}},
		{dir + "/su", []string{dir + "/sub/"}},
		{"zzz", nil},
	}
	for _, tt := range tests {
		got, err := listFiles(tt.prefix, dir, home)
		if err != nil {
			t.Errorf("listFiles(%q): %v", tt.prefix, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("listFiles(%q) = %q, want %q", tt.prefix, got, tt.want)
		}
	}

	if _, err := listFiles("missing/", dir, home); err == nil {
		t.Error("listFiles(missing/): want an error")
	}
}

// writeFiles creates each file (slash-separated, relative to dir) with its contents.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestIsMarkdown(t *testing.T) {
	for name, want := range map[string]bool{
		"a.md": true, "A.MD": true, "a.markdown": true, "a.mdown": true, "a.mkd": true,
		"a.txt": true, "a.png": false, "md": false, "a.md.bak": false, "": false,
	} {
		if got := isMarkdown(name); got != want {
			t.Errorf("isMarkdown(%q) = %v, want %v", name, got, want)
		}
	}
	if got, want := markdownFilter[0].Pattern, "*.md;*.markdown;*.mdown;*.mkd;*.txt"; got != want {
		t.Errorf("markdown filter pattern = %q, want %q", got, want)
	}
}

func TestResolveRelative(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	writeFiles(t, dir, map[string]string{
		"doc.md": "", "other.md": "", "my notes.md": "", "sub/deep.md": "", "home/h.md": "",
	})
	t.Setenv("HOME", home)

	a := NewApp("", false)
	if _, err := a.ResolveRelative("other.md"); err == nil {
		t.Error("with no file open: want an error")
	}

	a = NewApp(filepath.Join(dir, "doc.md"), false)
	tests := []struct{ href, want string }{
		{"other.md", "other.md"},
		{"other.md#intro", "other.md"},
		{"my%20notes.md", "my notes.md"},
		{"sub/deep.md", "sub/deep.md"},
		{"./sub/../other.md", "other.md"},
		{filepath.Join(dir, "other.md"), "other.md"},
		{"~/h.md", "home/h.md"},
	}
	for _, tt := range tests {
		got, err := a.ResolveRelative(tt.href)
		if err != nil {
			t.Errorf("ResolveRelative(%q): %v", tt.href, err)
			continue
		}
		if want := filepath.Join(dir, filepath.FromSlash(tt.want)); got != want {
			t.Errorf("ResolveRelative(%q) = %q, want %q", tt.href, got, want)
		}
	}
	for _, href := range []string{"missing.md", "%zz.md"} {
		if got, err := a.ResolveRelative(href); err == nil {
			t.Errorf("ResolveRelative(%q) = %q, want an error", href, got)
		}
	}
}

func TestGetDocument(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfg)
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"doc.md": "# Title\n\n<span onclick=\"x()\">hi</span>\n",
	})
	doc := filepath.Join(dir, "doc.md")

	t.Run("no file", func(t *testing.T) {
		d := NewApp("", false).GetDocument()
		if d.Path != "" || d.HTML != "" || d.Error != "" {
			t.Errorf("got %+v, want only settings", d)
		}
		if d.Theme != "dark" || d.TOC != "overlay" || d.CSS == "" || d.CodeCSS.Dark == "" || d.CodeCSS.Light == "" {
			t.Errorf("defaults missing: theme %q toc %q", d.Theme, d.TOC)
		}
	})

	t.Run("sanitized", func(t *testing.T) {
		d := NewApp(doc, false).GetDocument()
		if d.Error != "" || d.ConfigError != "" {
			t.Fatalf("errors: %q, %q", d.Error, d.ConfigError)
		}
		if d.Name != "doc.md" || d.Path != doc {
			t.Errorf("name %q path %q", d.Name, d.Path)
		}
		if !strings.Contains(d.HTML, "Title</h1>") {
			t.Errorf("HTML missing heading: %s", d.HTML)
		}
		if strings.Contains(d.HTML, "onclick") {
			t.Errorf("onclick not removed: %s", d.HTML)
		}
	})

	t.Run("-html", func(t *testing.T) {
		if d := NewApp(doc, true).GetDocument(); !strings.Contains(d.HTML, "onclick") {
			t.Errorf("onclick removed with -html: %s", d.HTML)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if d := NewApp(filepath.Join(dir, "gone.md"), false).GetDocument(); d.Error == "" {
			t.Error("want an error for a missing file")
		}
	})

	t.Run("config", func(t *testing.T) {
		writeFiles(t, cfg, map[string]string{
			"katdown/config.toml": "theme = \"light\"\n[toc]\nstyle = \"nope\"\n[html]\nunsafe = true\n",
			"katdown/style.css":   "body { color: red }",
		})
		d := NewApp(doc, false).GetDocument()
		if d.Theme != "light" {
			t.Errorf("theme %q, want light", d.Theme)
		}
		if d.TOC != "overlay" || !strings.Contains(d.ConfigError, "toc.style") {
			t.Errorf("bad toc.style: toc %q, config error %q", d.TOC, d.ConfigError)
		}
		if !d.CustomCSS || !strings.HasPrefix(d.CSS, "body { color: red }") {
			t.Errorf("custom CSS not used: %v %.40q", d.CustomCSS, d.CSS)
		}
		if !strings.Contains(d.HTML, "onclick") {
			t.Errorf("unsafe = true didn't keep onclick: %s", d.HTML)
		}
	})
}

func TestLocalFiles(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"docs/doc.md":        "",
		"docs/img.png":       "png",
		"docs/sub/pic.png":   "sub",
		"docs/two words.txt": "spaced",
		"secret.txt":         "secret",
	})
	a := NewApp(filepath.Join(root, "docs", "doc.md"), false)

	tests := []struct {
		method, target string
		status         int
		body           string
	}{
		{"GET", "/img.png", 200, "png"},
		{"GET", "/sub/pic.png", 200, "sub"},
		{"GET", "/two%20words.txt", 200, "spaced"},
		{"GET", "/missing.png", 404, ""},
		{"GET", "/../secret.txt", 400, ""}, // http.ServeFile refuses ".." itself
		{"GET", "/%2e%2e/secret.txt", 400, ""},
		{"GET", "/sub/../../secret.txt", 400, ""},
		{"POST", "/img.png", 404, ""},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(tt.method, "/", nil)
		// Set the path directly so httptest doesn't clean it first.
		r.URL.Path, _ = url.PathUnescape(tt.target)
		w := httptest.NewRecorder()
		a.localFiles(w, r)
		if w.Code != tt.status {
			t.Errorf("%s %s: status %d, want %d", tt.method, tt.target, w.Code, tt.status)
		}
		if tt.status == 200 && w.Body.String() != tt.body {
			t.Errorf("%s %s: body %q, want %q", tt.method, tt.target, w.Body.String(), tt.body)
		}
		if strings.Contains(w.Body.String(), "secret") {
			t.Errorf("%s %s: served a file outside the document folder", tt.method, tt.target)
		}
	}

	w := httptest.NewRecorder()
	NewApp("", false).localFiles(w, httptest.NewRequest("GET", "/img.png", nil))
	if w.Code != 404 {
		t.Errorf("with no file open: status %d, want 404", w.Code)
	}
}

func TestWithFallback(t *testing.T) {
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "fallback")
		io.WriteString(w, "fallback")
	})
	tests := []struct {
		name   string
		h      http.HandlerFunc
		status int
		body   string
		from   string
	}{
		{"found", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-From", "assets")
			w.WriteHeader(200)
			io.WriteString(w, "asset")
		}, 200, "asset", "assets"},
		{"write without WriteHeader", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "asset")
		}, 200, "asset", ""},
		{"other status", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusForbidden)
		}, 403, "nope\n", ""},
		{"not found", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-From", "assets")
			http.NotFound(w, r)
		}, 200, "fallback", "fallback"},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		withFallback(tt.h, fallback).ServeHTTP(w, httptest.NewRequest("GET", "/x", nil))
		if w.Code != tt.status || w.Body.String() != tt.body || w.Header().Get("X-From") != tt.from {
			t.Errorf("%s: got %d %q from %q, want %d %q from %q", tt.name,
				w.Code, w.Body.String(), w.Header().Get("X-From"), tt.status, tt.body, tt.from)
		}
	}
}

func TestMediaRoute(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	writeFiles(t, root, map[string]string{
		"notes/journal/day.md":       "",
		"notes/journal/here.png":     "here",
		"notes/assets/up.png":        "up",
		"notes/assets/two words.jpg": "spaced",
		"notes/secret.txt":           "secret",
		"home/pics/me.gif":           "home",
		"notes/dir.png/x":            "",
	})
	dir := filepath.Join(root, "notes", "journal")
	abs := filepath.ToSlash(filepath.Join(root, "notes", "assets", "up.png"))

	tests := []struct {
		src    string
		status int
		body   string
	}{
		{"here.png", 200, "here"},
		{"../assets/up.png", 200, "up"},
		{"../assets/two%20words.jpg", 200, "spaced"},
		{"../assets/up.png?v=2#x", 200, "up"},
		{"~/pics/me.gif", 200, "home"},
		{abs, 200, "up"},
		{"file://" + abs, 200, "up"},
		{"../secret.txt", 404, ""},
		{"../dir.png", 404, ""},
		{"missing.png", 404, ""},
		{"", 404, ""},
	}
	for _, tt := range tests {
		r := httptest.NewRequest("GET", mediaRoute+"?src="+url.QueryEscape(tt.src), nil)
		w := httptest.NewRecorder()
		serveMedia(w, r, dir, home)
		if w.Code != tt.status {
			t.Errorf("%q: status %d, want %d", tt.src, w.Code, tt.status)
		}
		if tt.status == 200 && w.Body.String() != tt.body {
			t.Errorf("%q: body %q, want %q", tt.src, w.Body.String(), tt.body)
		}
	}

	// localFiles routes mediaRoute to serveMedia.
	a := NewApp(filepath.Join(dir, "day.md"), false)
	w := httptest.NewRecorder()
	a.localFiles(w, httptest.NewRequest("GET", mediaRoute+"?src=../assets/up.png", nil))
	if w.Code != 200 || w.Body.String() != "up" {
		t.Errorf("localFiles %s: %d %q", mediaRoute, w.Code, w.Body.String())
	}
}
