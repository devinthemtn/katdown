package recent

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestAddList(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "state", "recent.json")
	touch := func(name string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	a, b, c := touch("a.md"), touch("b.md"), touch("c.md")

	if got := List(file, ""); got != nil {
		t.Errorf("empty list = %q", got)
	}
	for _, p := range []string{a, b, c, a} {
		if err := Add(file, p); err != nil {
			t.Fatal(err)
		}
	}
	if got, want := List(file, ""), []string{a, c, b}; !reflect.DeepEqual(got, want) {
		t.Errorf("List = %q, want %q", got, want)
	}
	if got, want := List(file, a), []string{c, b}; !reflect.DeepEqual(got, want) {
		t.Errorf("List skipping a = %q, want %q", got, want)
	}
	os.Remove(c)
	if got, want := List(file, ""), []string{a, b}; !reflect.DeepEqual(got, want) {
		t.Errorf("List after removing c = %q, want %q", got, want)
	}

	if err := os.WriteFile(file, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := List(file, ""); got != nil {
		t.Errorf("damaged list = %q", got)
	}
}

func TestMax(t *testing.T) {
	file := filepath.Join(t.TempDir(), "recent.json")
	for i := range Max + 5 {
		if err := Add(file, "/f"+strconv.Itoa(i)); err != nil {
			t.Fatal(err)
		}
	}
	if got := load(file); len(got) != Max || got[0] != "/f"+strconv.Itoa(Max+4) {
		t.Errorf("kept %d, first %q", len(got), got[0])
	}
}
