// Package editor works out how to open a file in the user's editor.
package editor

import (
	"fmt"
	"os/exec"
	"strings"
)

// Env is what Command needs to know about the system, so tests can fake it.
type Env struct {
	Getenv   func(string) string
	LookPath func(string) (string, error)
	GOOS     string
	WSL      bool
}

// Command returns the command that opens path in an editor. The first of
// these is used: setting (the config's editor.command), $VISUAL, then the OS
// default. $EDITOR is skipped: by convention it is a terminal editor, and
// katdown may have no terminal to run it in.
func Command(setting, path string, env Env) (*exec.Cmd, error) {
	cmd := strings.TrimSpace(setting)
	if cmd == "" {
		cmd = strings.TrimSpace(env.Getenv("VISUAL"))
	}
	if cmd != "" {
		return userCommand(cmd, path, env)
	}

	switch env.GOOS {
	case "windows":
		return exec.Command("notepad.exe", path), nil
	case "darwin":
		return exec.Command("open", "-t", path), nil
	}
	if _, err := env.LookPath("xdg-open"); err == nil || !env.WSL {
		return exec.Command("xdg-open", path), nil
	}
	// WSL without xdg-open. Not explorer.exe: it would hand a .md file back
	// to katdown if katdown is registered for .md on the Windows side.
	return exec.Command("sh", "-c", `notepad.exe "$(wslpath -w "$1")"`, "katdown", path), nil
}

// userCommand runs a command line such as "code --wait" with path appended.
func userCommand(cmd, path string, env Env) (*exec.Cmd, error) {
	fields := strings.Fields(cmd)
	if _, err := env.LookPath(fields[0]); err != nil {
		return nil, fmt.Errorf("editor %q not found", fields[0])
	}
	if env.GOOS == "windows" {
		return exec.Command(fields[0], append(fields[1:], path)...), nil
	}
	// Through the shell, as git runs $VISUAL, so the user's quoting is kept.
	return exec.Command("sh", "-c", cmd+` "$1"`, "katdown", path), nil
}
