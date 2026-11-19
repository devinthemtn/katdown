package main

import (
	"os"
	"path/filepath"
	"reflect"
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
