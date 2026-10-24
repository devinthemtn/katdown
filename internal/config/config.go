// Package config locates katdown's config folder and resolves the stylesheet
// used to display documents.
package config

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

//go:generate go run ./chromagen chroma.css

// AppName is used for the config folder name.
const AppName = "katdown"

// StyleFile is the name of the user's custom stylesheet inside Dir().
const StyleFile = "style.css"

var (
	//go:embed default.css
	baseCSS string
	//go:embed chroma.css
	chromaCSS string
)

// DefaultCSS is the built-in stylesheet, written out by MakeConfig.
var DefaultCSS = baseCSS + "\n" + chromaCSS

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

// MakeConfig creates the config folder and writes DefaultCSS to style.css.
// An existing style.css is left untouched and created is false.
func MakeConfig() (path string, created bool, err error) {
	path, err = CSSPath()
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return path, false, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return path, false, nil
	}
	if err != nil {
		return path, false, err
	}
	if _, err := f.WriteString(DefaultCSS); err != nil {
		f.Close()
		return path, false, err
	}
	return path, true, f.Close()
}
