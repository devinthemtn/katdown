package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileNotifiesOnWriteAndRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}

	changed := make(chan struct{}, 10)
	w, err := File(path, func() { changed <- struct{}{} })
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	wait := func(what string) {
		t.Helper()
		select {
		case <-changed:
		case <-time.After(2 * time.Second):
			t.Fatalf("no change reported after %s", what)
		}
	}

	if err := os.WriteFile(path, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	wait("write")

	// Editor-style save: write a temp file, rename it over the original.
	tmp := filepath.Join(dir, ".doc.md.swp")
	if err := os.WriteFile(tmp, []byte("c"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
	wait("rename")

	// Other files in the directory are ignored.
	if err := os.WriteFile(filepath.Join(dir, "other.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-changed:
		t.Fatal("change reported for an unrelated file")
	case <-time.After(400 * time.Millisecond):
	}
}
