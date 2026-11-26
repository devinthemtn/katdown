package render

import (
	"regexp"
	"strings"
	"testing"
)

func renderWith(t *testing.T, src string, opts Options) string {
	t.Helper()
	out, err := Render([]byte(src), opts)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCalloutBuiltin(t *testing.T) {
	out := render(t, "> [!WARNING]\n> Be **careful**.\n")
	for _, want := range []string{
		`<div class="callout callout-warning">`,
		`<p class="callout-title"><span class="callout-icon">⚠️</span>Warning</p>`,
		"<p>Be <strong>careful</strong>.</p>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "[!WARNING]") || strings.Contains(out, "blockquote") {
		t.Errorf("marker or blockquote left:\n%s", out)
	}
}

func TestCalloutTitleAndCase(t *testing.T) {
	out := render(t, "> [!note] Heads *up*\n> body\n")
	if !strings.Contains(out, `callout-note`) || !strings.Contains(out, "</span>Heads <em>up</em></p>") {
		t.Errorf("custom title:\n%s", out)
	}
}

func TestCalloutUnknownAndEscaped(t *testing.T) {
	for _, src := range []string{"> [!idea]\n> body\n", "> \\[!NOTE]\n> body\n", "> just a quote\n", "> text\n> [!NOTE]\n"} {
		if out := render(t, src); !strings.Contains(out, "<blockquote>") || strings.Contains(out, "callout") {
			t.Errorf("%q should stay a blockquote:\n%s", src, out)
		}
	}
}

func TestCalloutCustom(t *testing.T) {
	types := CalloutTypes(map[string]Callout{"idea": {Icon: "🧠", Title: "Big <idea>"}, "note": {Title: "FYI"}})
	out := renderWith(t, "> [!IDEA]\n> - one\n> - two\n\n> [!NOTE]\n", Options{Callouts: types})
	for _, want := range []string{
		`<div class="callout callout-idea">`,
		"🧠</span>Big &lt;idea&gt;</p>",
		"<li>one</li>",
		"ℹ️</span>FYI</p>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestCalloutRawHTMLAndHighlight(t *testing.T) {
	braces := []*regexp.Regexp{regexp.MustCompile(`\{[^{}]+\}`)}
	for _, raw := range []bool{false, true} {
		out := renderWith(t, "> [!TIP]\n> {x}\n", Options{Highlights: braces, RawHTML: raw})
		if !strings.Contains(out, `<div class="callout callout-tip">`) || !strings.Contains(out, "katdown-highlight") {
			t.Errorf("raw=%v:\n%s", raw, out)
		}
	}
}
