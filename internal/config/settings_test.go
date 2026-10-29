package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestDefaultSettingsFileMatchesDefaults(t *testing.T) {
	// Decode into zero values so a key missing from default.toml shows up.
	var s Settings
	if _, err := toml.Decode(DefaultSettingsFile, &s); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, Defaults()) {
		t.Errorf("default.toml = %+v\nDefaults() = %+v", s, Defaults())
	}
}

func TestLoadSettingsMissingFile(t *testing.T) {
	useTempConfig(t)
	s, err := LoadSettings()
	if err != nil || !reflect.DeepEqual(s, Defaults()) {
		t.Errorf("LoadSettings() = %+v, %v; want defaults, nil", s, err)
	}
}

func TestLoadSettingsPartial(t *testing.T) {
	writeSettings(t, "theme = \"light\"\n[font]\nsize = 18\n")
	s, err := LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	want := Defaults()
	want.Theme = "light"
	want.Font.Size = 18
	if !reflect.DeepEqual(s, want) {
		t.Errorf("LoadSettings() = %+v; want %+v", s, want)
	}
}

func TestParseSettingsInvalidValues(t *testing.T) {
	s, err := ParseSettings(`
theme = "purple"
bogus = 1
[font]
body = "Evil\"; } body { color: red"
size = 2
[layout]
width = "calc(100% - 1px)"
[code]
dark = "no-such-style"
light = "monokai"
`)
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"theme", "bogus", "font.body", "font.size", "layout.width", "code.dark"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %s: %v", want, err)
		}
	}
	want := Defaults()
	want.Code.Light = "monokai" // the one valid setting still applies
	if !reflect.DeepEqual(s, want) {
		t.Errorf("ParseSettings() = %+v; want %+v", s, want)
	}
}

func TestParseSettingsSyntaxError(t *testing.T) {
	s, err := ParseSettings("theme = \"light\"\n[font\n")
	if err == nil || !strings.Contains(err.Error(), SettingsFile) {
		t.Errorf("expected a %s syntax error, got %v", SettingsFile, err)
	}
	if !reflect.DeepEqual(s, Defaults()) {
		t.Errorf("expected defaults on a syntax error, got %+v", s)
	}
}

func TestCSSVars(t *testing.T) {
	css := Defaults().CSSVars()
	for _, want := range []string{`--font-body: "Inter Variable", `, "--font-size: 16px;", "--content-width: 46rem;"} {
		if !strings.Contains(css, want) {
			t.Errorf("CSSVars() missing %q:\n%s", want, css)
		}
	}
}

func TestCodeCSS(t *testing.T) {
	dark, light := Defaults().CodeCSS()
	for _, css := range []string{dark, light} {
		if !strings.Contains(css, ".chroma .k") {
			t.Errorf("CodeCSS() missing token rules:\n%.200s", css)
		}
	}
	if dark == light {
		t.Error("dark and light code CSS are identical")
	}
}

func TestMakeSettingsKeepsExisting(t *testing.T) {
	useTempConfig(t)
	path, created, err := MakeSettings()
	if err != nil || !created {
		t.Fatalf("MakeSettings() = %q, %v, %v", path, created, err)
	}
	if data, _ := os.ReadFile(path); string(data) != DefaultSettingsFile {
		t.Error("config.toml does not match DefaultSettingsFile")
	}
	if err := os.WriteFile(path, []byte("theme = \"light\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, created, err := MakeSettings(); err != nil || created {
		t.Errorf("second MakeSettings() created = %v, err = %v", created, err)
	}
	if data, _ := os.ReadFile(path); string(data) != "theme = \"light\"\n" {
		t.Errorf("existing config.toml was overwritten: %q", data)
	}
}

func writeSettings(t *testing.T, data string) {
	t.Helper()
	path := filepath.Join(filepath.Dir(useTempConfig(t)), SettingsFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHighlightSettings(t *testing.T) {
	if res := Defaults().HighlightPatterns(); len(res) != 1 || !res[0].MatchString("say {hello}") {
		t.Errorf("default highlight patterns = %v; want one matching {hello}", res)
	}

	s, err := ParseSettings("[highlight]\npatterns = []\n")
	if err != nil || len(s.HighlightPatterns()) != 0 {
		t.Errorf("patterns = [] gave %v, %v; want none", s.Highlight.Patterns, err)
	}

	s, err = ParseSettings("[highlight]\npatterns = ['TODO', '(unclosed']\ncolor = \"orange\"\n")
	if err == nil || !strings.Contains(err.Error(), "highlight.patterns") {
		t.Errorf("expected an error for the bad pattern, got %v", err)
	}
	if !reflect.DeepEqual(s.Highlight.Patterns, []string{"TODO"}) {
		t.Errorf("patterns = %q; want only the valid one", s.Highlight.Patterns)
	}
	if css := s.CSSVars(); !strings.Contains(css, "--highlight: orange;") {
		t.Errorf("CSSVars() missing the highlight color:\n%s", css)
	}

	s, err = ParseSettings("[highlight]\ncolor = \"red; } body { display: none\"\n")
	if err == nil || s.Highlight.Color != "" {
		t.Errorf("unsafe color accepted: %q, %v", s.Highlight.Color, err)
	}
}
