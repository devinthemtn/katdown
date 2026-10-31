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
	"github.com/microcosm-cc/bluemonday"
	"go.abhg.dev/goldmark/mermaid"
)

// md passes raw HTML through; Render sanitizes the result unless the user
// trusts the document. Both md and policy are safe for concurrent use.
var (
	md     = newMarkdown()
	policy = newPolicy()
)

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
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
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

// newPolicy allows the HTML that READMEs commonly use plus what katdown's own
// output needs, and nothing that runs script: the page shares a webview with
// the Go bindings.
func newPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.RequireNoFollowOnLinks(false) // no search engine reads these pages
	// chroma, mermaid, math, highlight and footnote classes; heading IDs.
	p.AllowAttrs("class", "id").Globally()
	// GFM task lists.
	p.AllowAttrs("type", "checked", "disabled").OnElements("input")
	p.AllowElements("details", "summary", "kbd", "sub", "sup", "mark", "abbr")
	p.AllowAttrs("open").OnElements("details")
	p.AllowAttrs("title").OnElements("abbr")
	p.AllowAttrs("align").OnElements("p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "img", "td", "th")
	// GFM table column alignment is the only inline style let through.
	p.AllowStyles("text-align").MatchingEnum("left", "center", "right").OnElements("td", "th")
	return p
}

// Render returns src rendered as an HTML fragment. Text matching any of the
// highlight patterns is wrapped in <span class="katdown-highlight">. Unless
// rawHTML is set, the output is sanitized: safe raw HTML such as <details> or
// <kbd> is kept, while scripts, styles, event handlers and javascript: URLs
// are removed.
func Render(src []byte, highlights []*regexp.Regexp, rawHTML bool) (string, error) {
	pc := parser.NewContext()
	pc.Set(highlightKey, highlights)
	var buf bytes.Buffer
	if err := md.Convert(src, &buf, parser.WithContext(pc)); err != nil {
		return "", err
	}
	if rawHTML {
		return buf.String(), nil
	}
	return policy.SanitizeReader(&buf).String(), nil
}
