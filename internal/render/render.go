// Package render converts markdown to the HTML shown in the viewer.
package render

import (
	"bytes"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/mermaid"
)

// md is safe for concurrent use. Raw HTML in documents is escaped (goldmark's
// default) because the page shares a webview with the Go bindings.
var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
		// Mermaid's transformer claims ```mermaid fences before highlighting
		// sees them; the frontend draws the diagrams, which ship with the app.
		&mermaid.Extender{RenderMode: mermaid.RenderModeClient, NoScript: true},
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
		),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

// Render returns src rendered as an HTML fragment.
func Render(src []byte) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
