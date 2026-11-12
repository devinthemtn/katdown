package config

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// uncomment turns every commented-out setting back on, but not table
// headers: added headers must already be on, so a setting means what its
// place in the file says.
func uncomment(data string) string {
	re := regexp.MustCompile(`(?m)^# ([A-Za-z0-9_.-]+ = )`)
	return re.ReplaceAllString(data, "$1")
}

func TestUpdateSettingsTextRestoresDefaults(t *testing.T) {
	for _, data := range []string{
		"",
		"theme = \"light\"\n",
		"[font]\nsize = 18\n\n[window]\nwidth = 800\n",
		"# my settings\n\n[toc]\nstyle = \"sidebar\"\n",
		"font.size = 18\n[code]\ndark = \"monokai\"\n",
	} {
		out, added, err := UpdateSettingsText(data)
		if err != nil {
			t.Fatalf("%q: %v", data, err)
		}
		if len(added) == 0 {
			t.Errorf("%q: nothing added", data)
		}
		// Uncommenting what was added gives a complete file: every setting
		// is defined, and the user's own values are kept.
		s, err := ParseSettings(uncomment(out))
		if err != nil {
			t.Fatalf("%q: uncommented output does not parse: %v\n%s", data, err, out)
		}
		user, _ := ParseSettings(data)
		if !reflect.DeepEqual(s, user) {
			t.Errorf("%q: uncommented output = %+v; want %+v\n%s", data, s, user, out)
		}
		// Nothing is left missing.
		if _, again, _ := UpdateSettingsText(uncomment(out)); len(again) != 0 {
			t.Errorf("%q: still missing %v after uncommenting\n%s", data, again, out)
		}
	}
}

func TestUpdateSettingsTextIdempotent(t *testing.T) {
	first, _, err := UpdateSettingsText("[font]\nsize = 18\n")
	if err != nil {
		t.Fatal(err)
	}
	second, added, err := UpdateSettingsText(first)
	if err != nil || len(added) != 0 || second != first {
		t.Errorf("second update added %v, err %v, changed = %v", added, err, second != first)
	}
}

func TestUpdateSettingsTextPlacement(t *testing.T) {
	out, added, err := UpdateSettingsText("[font]\nsize = 18\n\n# window stuff\n[window]\nwidth = 800\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"theme", "font.body", "window.height", "toc.style"} {
		found := false
		for _, a := range added {
			found = found || a == want
		}
		if !found {
			t.Errorf("added = %v; missing %s", added, want)
		}
	}
	font := strings.Index(out, "[font]")
	body := strings.Index(out, `# body = "Inter Variable"`)
	window := strings.Index(out, "# window stuff")
	if !(font < body && body < window) {
		t.Errorf("font.body not inside [font], above the window comment:\n%s", out)
	}
	if !strings.HasPrefix(out, "# dark, light, or system") {
		t.Errorf("theme not added at the top:\n%s", out)
	}
	if !strings.Contains(out, "\n[toc]\n") || strings.Contains(out, "# [toc]") {
		t.Errorf("missing [toc] table not appended:\n%s", out)
	}
}

// Uncommenting only the one setting under an appended table must set it
// there, not in the table above.
func TestUpdateSettingsTextAppendedTableHeader(t *testing.T) {
	out, _, err := UpdateSettingsText("[editor]\ncommand = \"\"\n")
	if err != nil {
		t.Fatal(err)
	}
	out = strings.Replace(out, `# style = "overlay"`, `style = "sidebar"`, 1)
	s, err := ParseSettings(out)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if s.TOC.Style != "sidebar" {
		t.Errorf("toc.style = %q; want sidebar\n%s", s.TOC.Style, out)
	}
}

func TestUpdateSettingsTextComplete(t *testing.T) {
	out, added, err := UpdateSettingsText(DefaultSettingsFile)
	if err != nil || len(added) != 0 || out != DefaultSettingsFile {
		t.Errorf("default file: added %v, err %v", added, err)
	}
}

func TestUpdateSettingsTextSyntaxError(t *testing.T) {
	if _, _, err := UpdateSettingsText("theme = \n"); err == nil {
		t.Error("expected an error")
	}
}

func TestUpdateSettingsFile(t *testing.T) {
	writeSettings(t, "theme = \"light\"\n")
	path, added, created, err := UpdateSettings()
	if err != nil || created || len(added) == 0 {
		t.Fatalf("UpdateSettings() = %v, %v, %v", added, created, err)
	}
	data, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(data), "theme = \"light\"\n") || !strings.Contains(string(data), "\n[toc]\n") {
		t.Errorf("config.toml after update:\n%s", data)
	}
}

func TestUpdateSettingsMissingFile(t *testing.T) {
	useTempConfig(t)
	path, _, created, err := UpdateSettings()
	if err != nil || !created {
		t.Fatalf("UpdateSettings() created = %v, err = %v", created, err)
	}
	if data, _ := os.ReadFile(path); string(data) != DefaultSettingsFile {
		t.Error("missing config.toml was not created with the defaults")
	}
}
