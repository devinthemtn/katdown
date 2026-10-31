// Package render converts markdown to the HTML shown in the viewer.
package render

import (
	"bytes"
	"regexp"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
)

// md and mdRaw are safe for concurrent use. md drops raw HTML in documents
// (goldmark's default) because the page shares a webview with the Go
// bindings; mdRaw passes it through, for documents the user trusts.
var (
	md    = newMarkdown()
	mdRaw = newMarkdown(goldmark.WithRendererOptions(html.WithUnsafe()))
)

func newMarkdown(opts ...goldmark.Option) goldmark.Markdown {
	return goldmark.New(append([]goldmark.Option{
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			math{},
			highlight{},
			// Mermaid's transformer claims ```mermaid fences before highlighting
			// sees them; the frontend draws the diagrams, which ship with the app.
			&mermaid.Extender{RenderMode: mermaid.RenderModeClient, NoScript: true},
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	}, opts...)...)
}

// Render returns src rendered as an HTML fragment. Text matching any of the
// highlight patterns is wrapped in <span class="katdown-highlight">. Raw HTML
// in src is dropped unless rawHTML is set.
func Render(src []byte, highlights []*regexp.Regexp, rawHTML bool) (string, error) {
	pc := parser.NewContext()
	pc.Set(highlightKey, highlights)
	var buf bytes.Buffer
	conv := md
	if rawHTML {
		conv = mdRaw
	}
	if err := conv.Convert(src, &buf, parser.WithContext(pc)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
