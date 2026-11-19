// Package recent keeps the list of files katdown has opened, newest first,
// for the :recent command.
package recent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
)

// Max is how many files the list keeps.
const Max = 30

var mu sync.Mutex

// Path returns where the list is stored: $XDG_STATE_HOME/katdown/recent.json
// (~/.local/state by default) on Linux, the config folder elsewhere.
func Path() (string, error) {
	if runtime.GOOS == "linux" {
		if dir := os.Getenv("XDG_STATE_HOME"); filepath.IsAbs(dir) {
			return filepath.Join(dir, "katdown", "recent.json"), nil
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "state", "katdown", "recent.json"), nil
		}
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "katdown", "recent.json"), nil
}

// Add moves p to the front of the list in file and saves it.
func Add(file, p string) error {
	mu.Lock()
	defer mu.Unlock()
	list := load(file)
	list = slices.DeleteFunc(list, func(x string) bool { return x == p })
	list = append([]string{p}, list...)
	if len(list) > Max {
		list = list[:Max]
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, file)
}

// List returns the files in file, newest first, leaving out skip and any
// that no longer exist.
func List(file, skip string) []string {
	mu.Lock()
	defer mu.Unlock()
	var out []string
	for _, p := range load(file) {
		if p == skip {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			out = append(out, p)
		}
	}
	return out
}

// load reads the list; a missing or damaged file is an empty list.
func load(file string) []string {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var list []string
	if json.Unmarshal(data, &list) != nil {
		return nil
	}
	return list
}
