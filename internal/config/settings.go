package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

// SettingsFile is the name of the settings file inside Dir().
const SettingsFile = "config.toml"

// DefaultSettingsFile is the commented settings file written by MakeSettings.
// Its values match Defaults().
//
//go:embed default.toml
var DefaultSettingsFile string

// Settings are the user's choices from config.toml.
type Settings struct {
	Theme  string       `toml:"theme"`
	Font   FontSettings `toml:"font"`
	Layout struct {
		Width string `toml:"width"`
	} `toml:"layout"`
	Code struct {
		Dark  string `toml:"dark"`
		Light string `toml:"light"`
	} `toml:"code"`
	Window struct {
		Width  int `toml:"width"`
		Height int `toml:"height"`
	} `toml:"window"`
	Highlight struct {
		Patterns []string `toml:"patterns"` // Go regular expressions
		Color    string   `toml:"color"`    // empty: the theme's --highlight
	} `toml:"highlight"`
	Editor struct {
		Command string `toml:"command"` // empty: $VISUAL, then the OS default
	} `toml:"editor"`
}

type FontSettings struct {
	Body string `toml:"body"`
	Mono string `toml:"mono"`
	Size int    `toml:"size"`
}

// Fallbacks follow the configured font, as in the default stylesheet.
const (
	bodyFallback = `-apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", sans-serif`
	monoFallback = `ui-monospace, SFMono-Regular, Menlo, Consolas, monospace`
)

var (
	cssLength = regexp.MustCompile(`^\d+(\.\d+)?(rem|em|px|ch|%)$`)
	cssColor  = regexp.MustCompile(`^(#[0-9a-fA-F]{3,8}|[a-zA-Z]+|(rgb|rgba|hsl|hsla)\([0-9.,%/ ]+\))$`)
	// Font names end up inside a quoted CSS string.
	badFontChars = `"';{}<>\`
)

// Defaults returns the built-in settings.
func Defaults() Settings {
	var s Settings
	s.Theme = "dark"
	s.Font = FontSettings{Body: "Inter Variable", Mono: "JetBrains Mono Variable", Size: 16}
	s.Layout.Width = "46rem"
	s.Code.Dark = "tokyonight-night"
	s.Code.Light = "github"
	s.Window.Width = 960
	s.Window.Height = 900
	s.Highlight.Patterns = []string{`\{[^{}]+\}`}
	return s
}

// SettingsPath returns the path of the settings file, whether or not it exists.
func SettingsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SettingsFile), nil
}

// LoadSettings reads config.toml. A missing file gives Defaults() and no
// error. On a syntax error the defaults are returned with the error; invalid
// or unknown settings are reported in the error and the rest still apply.
func LoadSettings() (Settings, error) {
	path, err := SettingsPath()
	if err != nil {
		return Defaults(), nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Defaults(), nil
	}
	if err != nil {
		return Defaults(), err
	}
	return ParseSettings(string(data))
}

// ParseSettings decodes settings from TOML; see LoadSettings.
func ParseSettings(data string) (Settings, error) {
	s := Defaults()
	meta, err := toml.Decode(data, &s)
	if err != nil {
		return Defaults(), fmt.Errorf("%s: %w", SettingsFile, err)
	}

	var errs []error
	for _, key := range meta.Undecoded() {
		errs = append(errs, fmt.Errorf("unknown setting %q", key.String()))
	}
	errs = append(errs, s.validate()...)
	if len(errs) > 0 {
		return s, fmt.Errorf("%s: %w", SettingsFile, errors.Join(errs...))
	}
	return s, nil
}

// validate resets each invalid setting to its default and reports it.
func (s *Settings) validate() []error {
	d := Defaults()
	var errs []error
	bad := func(name string, value any, want string) {
		errs = append(errs, fmt.Errorf("%s = %v: %s", name, value, want))
	}

	switch s.Theme {
	case "dark", "light", "system":
	default:
		bad("theme", fmt.Sprintf("%q", s.Theme), `want "dark", "light" or "system"`)
		s.Theme = d.Theme
	}
	if s.Font.Body == "" || strings.ContainsAny(s.Font.Body, badFontChars) {
		bad("font.body", fmt.Sprintf("%q", s.Font.Body), "not a usable font name")
		s.Font.Body = d.Font.Body
	}
	if s.Font.Mono == "" || strings.ContainsAny(s.Font.Mono, badFontChars) {
		bad("font.mono", fmt.Sprintf("%q", s.Font.Mono), "not a usable font name")
		s.Font.Mono = d.Font.Mono
	}
	if s.Font.Size < 8 || s.Font.Size > 40 {
		bad("font.size", s.Font.Size, "want 8 to 40")
		s.Font.Size = d.Font.Size
	}
	if !cssLength.MatchString(s.Layout.Width) {
		bad("layout.width", fmt.Sprintf("%q", s.Layout.Width), `want a CSS length like "46rem" or "800px"`)
		s.Layout.Width = d.Layout.Width
	}
	if _, ok := styles.Registry[s.Code.Dark]; !ok {
		bad("code.dark", fmt.Sprintf("%q", s.Code.Dark), "unknown chroma style")
		s.Code.Dark = d.Code.Dark
	}
	if _, ok := styles.Registry[s.Code.Light]; !ok {
		bad("code.light", fmt.Sprintf("%q", s.Code.Light), "unknown chroma style")
		s.Code.Light = d.Code.Light
	}
	if s.Window.Width < 300 || s.Window.Width > 10000 {
		bad("window.width", s.Window.Width, "want 300 to 10000")
		s.Window.Width = d.Window.Width
	}
	if s.Window.Height < 200 || s.Window.Height > 10000 {
		bad("window.height", s.Window.Height, "want 200 to 10000")
		s.Window.Height = d.Window.Height
	}
	var patterns []string
	for _, p := range s.Highlight.Patterns {
		if _, err := regexp.Compile(p); err != nil {
			bad("highlight.patterns", fmt.Sprintf("%q", p), err.Error())
			continue
		}
		patterns = append(patterns, p)
	}
	s.Highlight.Patterns = patterns
	if c := s.Highlight.Color; c != "" && !cssColor.MatchString(c) {
		bad("highlight.color", fmt.Sprintf("%q", c), `want a CSS color like "#e0af68" or "orange"`)
		s.Highlight.Color = d.Highlight.Color
	}
	return errs
}

// HighlightPatterns returns the compiled highlight patterns; invalid ones
// were already dropped by LoadSettings.
func (s Settings) HighlightPatterns() []*regexp.Regexp {
	var res []*regexp.Regexp
	for _, p := range s.Highlight.Patterns {
		if re, err := regexp.Compile(p); err == nil {
			res = append(res, re)
		}
	}
	return res
}

// CSSVars returns a rule that sets the stylesheet variables these settings
// control. It goes after the stylesheet so it also applies to a custom one.
func (s Settings) CSSVars() string {
	css := fmt.Sprintf(":root {\n  --font-body: %q, %s;\n  --font-mono: %q, %s;\n  --font-size: %dpx;\n  --content-width: %s;\n",
		s.Font.Body, bodyFallback, s.Font.Mono, monoFallback, s.Font.Size, s.Layout.Width)
	if s.Highlight.Color != "" {
		// Matches the light palette's specificity, and comes later, so it wins.
		css += "}\n:root[data-theme] {\n  --highlight: " + s.Highlight.Color + ";\n"
	}
	return css + "}\n"
}

// CodeCSS returns the syntax-highlighting CSS for the dark and light themes.
func (s Settings) CodeCSS() (dark, light string) {
	return chromaCSS(s.Code.Dark), chromaCSS(s.Code.Light)
}

func chromaCSS(style string) string {
	var buf bytes.Buffer
	buf.WriteString("/* ---- Syntax highlighting (chroma " + style + ") ---- */\n")
	if err := html.New(html.WithClasses(true)).WriteCSS(&buf, styles.Get(style)); err != nil {
		return ""
	}
	return buf.String()
}
