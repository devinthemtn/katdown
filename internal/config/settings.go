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
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"

	"katdown/internal/render"
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
	HTML struct {
		Unsafe bool `toml:"unsafe"` // same as -html: don't sanitize raw HTML
	} `toml:"html"`
	TOC struct {
		Style string `toml:"style"` // "overlay" or "sidebar"
	} `toml:"toc"`
	// Callouts adds or restyles "> [!type]" blocks ([callouts.<type>]).
	Callouts map[string]CalloutSettings `toml:"callouts"`
}

// CalloutSettings style one callout type. A new type needs a Color; empty
// fields of a built-in type keep its defaults.
type CalloutSettings struct {
	Color string `toml:"color"` // CSS color
	Icon  string `toml:"icon"`
	Title string `toml:"title"`
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
	calloutName = regexp.MustCompile(`^[a-z0-9-]+$`)
	cssLength   = regexp.MustCompile(`^\d+(\.\d+)?(rem|em|px|ch|%)$`)
	cssColor    = regexp.MustCompile(`^(#[0-9a-fA-F]{3,8}|[a-zA-Z]+|(rgb|rgba|hsl|hsla)\([0-9.,%/ ]+\))$`)
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
	s.TOC.Style = "overlay"
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
	var callouts map[string]CalloutSettings
	for raw, c := range s.Callouts {
		name := strings.ToLower(strings.TrimSpace(raw))
		key := "callouts." + raw
		_, builtin := render.BuiltinCallouts[name]
		switch {
		case !calloutName.MatchString(name):
			bad(key, fmt.Sprintf("%q", raw), "a type name may only use a-z, 0-9 and -")
		case c.Color != "" && !cssColor.MatchString(c.Color):
			bad(key+".color", fmt.Sprintf("%q", c.Color), `want a CSS color like "#e0af68" or "orange"`)
		case c.Color == "" && !builtin:
			bad(key+".color", `""`, "a new callout type needs a color")
		default:
			if callouts == nil {
				callouts = map[string]CalloutSettings{}
			}
			callouts[name] = c
		}
	}
	s.Callouts = callouts
	if len(callouts) == 0 {
		s.Callouts = nil
	}
	switch s.TOC.Style {
	case "overlay", "sidebar":
	default:
		bad("toc.style", fmt.Sprintf("%q", s.TOC.Style), `want "overlay" or "sidebar"`)
		s.TOC.Style = d.TOC.Style
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

// CalloutTypes returns the callout types to render: the built-ins plus the
// configured ones, with their icons and titles.
func (s Settings) CalloutTypes() map[string]render.Callout {
	custom := make(map[string]render.Callout, len(s.Callouts))
	for k, c := range s.Callouts {
		custom[k] = render.Callout{Icon: c.Icon, Title: c.Title}
	}
	return render.CalloutTypes(custom)
}

// calloutColors are the built-in types' colors (GitHub's) per theme.
var calloutColors = []struct{ name, dark, light string }{
	{"note", "#4493f8", "#0969da"},
	{"tip", "#3fb950", "#1a7f37"},
	{"important", "#ab7df8", "#8250df"},
	{"warning", "#d29922", "#9a6700"},
	{"caution", "#f85149", "#cf222e"},
}

// CalloutCSS styles the callout boxes and sets each type's --callout-color.
// It comes after the stylesheet, so it works with a custom style.css too;
// the box rules use :where() so any rule in style.css overrides them.
func (s Settings) CalloutCSS() string {
	var b strings.Builder
	b.WriteString(`:where(.markdown-body .callout) {
  --callout-color: var(--accent);
  border-left: 4px solid var(--callout-color);
  background: color-mix(in srgb, var(--callout-color) 9%, transparent);
  border-radius: 0 6px 6px 0;
  margin: 1em 0;
  padding: 0.6em 1em;
}
:where(.markdown-body .callout > :first-child) { margin-top: 0; }
:where(.markdown-body .callout > :last-child) { margin-bottom: 0; }
:where(.markdown-body .callout-title) {
  color: var(--callout-color);
  font-weight: 600;
  display: flex;
  gap: 0.4em;
  align-items: baseline;
}
`)
	for _, c := range calloutColors {
		fmt.Fprintf(&b, ":root .callout-%s { --callout-color: %s; }\n", c.name, c.dark)
		fmt.Fprintf(&b, ":root[data-theme=\"light\"] .callout-%s { --callout-color: %s; }\n", c.name, c.light)
	}
	names := make([]string, 0, len(s.Callouts))
	for k := range s.Callouts {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		if c := s.Callouts[k].Color; c != "" {
			// Beats the built-in rules above in both themes.
			fmt.Fprintf(&b, ":root[data-theme] .callout-%s { --callout-color: %s; }\n", k, c)
		}
	}
	return b.String()
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
