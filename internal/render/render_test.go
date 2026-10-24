package render

import (
	"strings"
	"testing"
)

func render(t *testing.T, src string) string {
	t.Helper()
	out, err := Render([]byte(src))
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
