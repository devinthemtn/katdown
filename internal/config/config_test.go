package config

import (
	"os"
	"path/filepath"
	"testing"
)

func useTempConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return filepath.Join(dir, AppName, StyleFile)
}

func TestMakeConfig(t *testing.T) {
	want := useTempConfig(t)

	path, created, err := MakeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if path != want || !created {
		t.Fatalf("MakeConfig() = %q, %v; want %q, true", path, created, want)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != DefaultCSS {
		t.Error("style.css does not match DefaultCSS")
	}
}

func TestMakeConfigKeepsExisting(t *testing.T) {
	path := useTempConfig(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, created, err := MakeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Error("MakeConfig reported creating a file that already existed")
	}
	if data, _ := os.ReadFile(path); string(data) != "body{}" {
		t.Errorf("existing style.css was overwritten: %q", data)
	}
}

func TestLoadCSS(t *testing.T) {
	path := useTempConfig(t)

	if css, custom := LoadCSS(); custom || css != DefaultCSS {
		t.Error("expected DefaultCSS when no custom file exists")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("body{color:red}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if css, custom := LoadCSS(); !custom || css != "body{color:red}" {
		t.Errorf("LoadCSS() = %q, %v; want custom CSS", css, custom)
	}
}
