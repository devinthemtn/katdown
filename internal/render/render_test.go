package render

import (
	"regexp"
	"strings"
	"testing"
)

func render(t *testing.T, src string) string {
	t.Helper()
	out, err := Render([]byte(src), nil)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestRenderGFM(t *testing.T) {
	out := render(t, "| a | b |\n|---|---|\n| 1 | 2 |\n\n- [x] done\n- [ ] todo\n")
	for _, want := range []string{"<table>", "<td>1</td>", `type="checkbox"`, "checked"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderHighlighting(t *testing.T) {
	out := render(t, "```go\nfunc main() {}\n```\n")
	if !strings.Contains(out, `class="chroma"`) {
		t.Errorf("expected chroma classes:\n%s", out)
	}
	if strings.Contains(out, "style=") {
		t.Errorf("expected class-based highlighting, got inline styles:\n%s", out)
	}
}

func TestRenderEscapesRawHTML(t *testing.T) {
	out := render(t, "hi\n\n<script>alert(1)</script>\n")
	if strings.Contains(out, "<script>") {
		t.Errorf("raw script tag was not escaped:\n%s", out)
	}
}

func TestRenderMermaid(t *testing.T) {
	out := render(t, "```mermaid\ngraph TD\n  A --> B\n```\n")
	if !strings.Contains(out, `<pre class="mermaid">`) {
		t.Errorf("expected mermaid container:\n%s", out)
	}
	if !strings.Contains(out, "A --&gt; B") {
		t.Errorf("expected escaped diagram source:\n%s", out)
	}
	if strings.Contains(out, "chroma") || strings.Contains(out, "<script") {
		t.Errorf("mermaid block should have no highlighting or script:\n%s", out)
	}
}

func TestRenderMath(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		not       []string
	}{
		{"inline", "Energy $E = mc^2$ here.", []string{`<span class="math inline">E = mc^2</span>`}, nil},
		{"no emphasis", "$a_1 b_1$ and $x*y*z$", []string{`<span class="math inline">a_1 b_1</span>`, `<span class="math inline">x*y*z</span>`}, []string{"<em>"}},
		{"inline display", "so $$\\sum_i x_i$$ ok", []string{`<span class="math display">\sum_i x_i</span>`}, nil},
		{"block", "$$\n\\int_0^1 x\\,dx \\\\\ny\n$$\n", []string{"<div class=\"math display\">\\int_0^1 x\\,dx \\\\\ny</div>"}, []string{"<p>"}},
		{"one-line block", "$$ a + b $$\n\nafter", []string{`<div class="math display"> a + b </div>`, "<p>after</p>"}, nil},
		{"escaped", "$a<b$", []string{`<span class="math inline">a&lt;b</span>`}, []string{"<b"}},
		{"prices", "It costs $5 and $10.", []string{"It costs $5 and $10."}, []string{"math"}},
		{"space after opener", "a $ b$ c", nil, []string{"math"}},
		{"unclosed", "only $x here", []string{"only $x here"}, []string{"math"}},
		{"code span", "`$x$`", []string{"<code>$x$</code>"}, []string{"math"}},
		{"escaped dollar", `\$x$ y`, []string{"$x$ y"}, []string{"math"}},
		{"escaped dollar in math", `$\$5$`, []string{`<span class="math inline">\$5</span>`}, nil},
		{"code fence", "```\n$$\nx\n$$\n```\n", []string{"$$\nx\n$$"}, []string{"math"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(t, tt.src)
			for _, w := range tt.want {
				if !strings.Contains(out, w) {
					t.Errorf("output missing %q:\n%s", w, out)
				}
			}
			for _, n := range tt.not {
				if strings.Contains(out, n) {
					t.Errorf("output should not contain %q:\n%s", n, out)
				}
			}
		})
	}
}

func TestRenderHighlight(t *testing.T) {
	braces := []*regexp.Regexp{regexp.MustCompile(`\{[^{}]+\}`)}
	tests := []struct {
		name, src string
		want      []string
		not       []string
	}{
		{"words", "Fill in {name} and {date}.", []string{
			`Fill in <span class="katdown-highlight">{name}</span> and <span class="katdown-highlight">{date}</span>.`,
		}, nil},
		{"in emphasis and links", "*see {this}* and [a {link}](x.md)", []string{
			`<em>see <span class="katdown-highlight">{this}</span></em>`,
			`<a href="x.md">a <span class="katdown-highlight">{link}</span></a>`,
		}, nil},
		{"line breaks kept", "one {a}\ntwo", []string{"<span class=\"katdown-highlight\">{a}</span>\ntwo"}, nil},
		{"split text nodes", "Hi {user_name}, {a < b}", []string{
			`<span class="katdown-highlight">{user_name}</span>`,
			`<span class="katdown-highlight">{a &lt; b}</span>`,
		}, nil},
		{"hard break", "{a}  \nb", []string{`<span class="katdown-highlight">{a}</span><br>`}, nil},
		{"escaped", `not \{this}`, []string{"not {this}"}, []string{"katdown-highlight"}},
		{"code span", "`{x}`", []string{"<code>{x}</code>"}, []string{"katdown-highlight"}},
		{"code block", "```\n{x}\n```\n", nil, []string{"katdown-highlight"}},
		{"math", `$\frac{1}{2}$`, nil, []string{"katdown-highlight"}},
		{"empty braces", "a {} b", nil, []string{"katdown-highlight"}},
		{"escapes html", "{a < b & c}", []string{`<span class="katdown-highlight">{a &lt; b &amp; c}</span>`}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Render([]byte(tt.src), braces)
			if err != nil {
				t.Fatal(err)
			}
			for _, w := range tt.want {
				if !strings.Contains(out, w) {
					t.Errorf("output missing %q:\n%s", w, out)
				}
			}
			for _, n := range tt.not {
				if strings.Contains(out, n) {
					t.Errorf("output should not contain %q:\n%s", n, out)
				}
			}
		})
	}
}

func TestRenderHighlightOverlapping(t *testing.T) {
	res := []*regexp.Regexp{regexp.MustCompile(`TODO`), regexp.MustCompile(`TODO:? \w+`)}
	out, err := Render([]byte("x TODO: fix y"), res)
	if err != nil {
		t.Fatal(err)
	}
	if want := `x <span class="katdown-highlight">TODO: fix</span> y`; !strings.Contains(out, want) {
		t.Errorf("output missing %q:\n%s", want, out)
	}
}
