package editor

import (
	"errors"
	"reflect"
	"testing"
)

func fakeEnv(goos string, wsl bool, vars map[string]string, onPath ...string) Env {
	return Env{
		Getenv: func(k string) string { return vars[k] },
		LookPath: func(name string) (string, error) {
			for _, p := range onPath {
				if p == name {
					return "/usr/bin/" + name, nil
				}
			}
			return "", errors.New("not found")
		},
		GOOS: goos,
		WSL:  wsl,
	}
}

func TestCommand(t *testing.T) {
	const file = "/docs/a b.md"
	tests := []struct {
		name    string
		setting string
		env     Env
		want    []string
	}{
		{"setting beats VISUAL", "gvim -f", fakeEnv("linux", false, map[string]string{"VISUAL": "code"}, "gvim", "code"),
			[]string{"sh", "-c", `gvim -f "$1"`, "katdown", file}},
		{"VISUAL with args", "", fakeEnv("linux", false, map[string]string{"VISUAL": "code --wait"}, "code"),
			[]string{"sh", "-c", `code --wait "$1"`, "katdown", file}},
		{"EDITOR ignored", "", fakeEnv("linux", false, map[string]string{"EDITOR": "vim"}, "vim", "xdg-open"),
			[]string{"xdg-open", file}},
		{"linux default", "", fakeEnv("linux", false, nil), []string{"xdg-open", file}},
		{"wsl with xdg-open", "", fakeEnv("linux", true, nil, "xdg-open"), []string{"xdg-open", file}},
		{"wsl without xdg-open", "", fakeEnv("linux", true, nil),
			[]string{"sh", "-c", `notepad.exe "$(wslpath -w "$1")"`, "katdown", file}},
		{"windows default", "", fakeEnv("windows", false, nil), []string{"notepad.exe", file}},
		{"windows VISUAL", "", fakeEnv("windows", false, map[string]string{"VISUAL": "code --wait"}, "code"),
			[]string{"code", "--wait", file}},
		{"mac default", "", fakeEnv("darwin", false, nil), []string{"open", "-t", file}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Command(tt.setting, file, tt.env)
			if err != nil {
				t.Fatal(err)
			}
			got := cmd.Args
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("args = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCommandMissingEditor(t *testing.T) {
	_, err := Command("", "/a.md", fakeEnv("linux", false, map[string]string{"VISUAL": "nosuch"}))
	if err == nil {
		t.Error("expected an error for an editor that isn't installed")
	}
}
