// Package config locates katdown's config folder and resolves the settings
// and stylesheet used to display documents.
package config

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// AppName is used for the config folder name.
const AppName = "katdown"

// StyleFile is the name of the user's custom stylesheet inside Dir().
const StyleFile = "style.css"

// DefaultCSS is the built-in stylesheet, written out by MakeStyle. Syntax
// highlighting is not part of it; see Settings.CodeCSS.
//
//go:embed default.css
var DefaultCSS string

// Dir returns the katdown config folder, e.g. ~/.config/katdown on Linux.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppName), nil
}

// CSSPath returns the path of the custom stylesheet, whether or not it exists.
func CSSPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, StyleFile), nil
}

// LoadCSS returns the custom stylesheet if one exists, otherwise DefaultCSS.
// A custom stylesheet replaces the default entirely.
func LoadCSS() (css string, custom bool) {
	path, err := CSSPath()
	if err != nil {
		return DefaultCSS, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultCSS, false
	}
	return string(data), true
}

// MakeSettings creates the config folder and writes DefaultSettingsFile to
// config.toml. An existing config.toml is left untouched and created is false.
func MakeSettings() (path string, created bool, err error) {
	path, err = SettingsPath()
	if err != nil {
		return "", false, err
	}
	created, err = writeNew(path, DefaultSettingsFile)
	return path, created, err
}

// MakeStyle creates the config folder and writes DefaultCSS to style.css.
// An existing style.css is left untouched and created is false.
func MakeStyle() (path string, created bool, err error) {
	path, err = CSSPath()
	if err != nil {
		return "", false, err
	}
	created, err = writeNew(path, DefaultCSS)
	return path, created, err
}

// writeNew writes data to path unless the file already exists.
func writeNew(path, data string) (created bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := f.WriteString(data); err != nil {
		f.Close()
		return false, err
	}
	return true, f.Close()
}
