package render

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Callout is how one "> [!type]" block type is shown. The colour is not here:
// it comes from the stylesheet (see config.Settings.CalloutCSS).
type Callout struct {
	Icon  string // shown before the title
	Title string // used when the marker line has no title of its own
}

// BuiltinCallouts are the GitHub alert types, which work without config.
var BuiltinCallouts = map[string]Callout{
	"note":      {Icon: "ℹ️", Title: "Note"},
	"tip":       {Icon: "💡", Title: "Tip"},
	"important": {Icon: "❗", Title: "Important"},
	"warning":   {Icon: "⚠️", Title: "Warning"},
	"caution":   {Icon: "⛔", Title: "Caution"},
}

// callout turns a blockquote whose first line is [!type] into
// <div class="callout callout-type"> with an icon and title, like GitHub's
// alerts. Text after the marker replaces the title. Unknown types stay plain
// blockquotes. The known types are passed per Render through calloutKey.
type callout struct{}

var calloutKey = parser.NewContextKey()

// calloutAttr marks a blockquote as a callout of that type.
var calloutAttr = []byte("katdown-callout")

var calloutMarker = regexp.MustCompile(`^\[!([A-Za-z0-9-]+)\][ \t]*([^\n]*?)[ \t]*\r?\n?$`)

var kindCalloutTitle = ast.NewNodeKind("CalloutTitle")

// calloutTitle is the callout's first line. Its Lines, if any, are the custom
// title, which the inline parser turns into children; otherwise Label is used.
type calloutTitle struct {
	ast.BaseBlock
	Icon, Label string
}

func (n *calloutTitle) Kind() ast.NodeKind { return kindCalloutTitle }

func (n *calloutTitle) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"Icon": n.Icon, "Label": n.Label}, nil)
}

func (callout) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithParagraphTransformers(util.Prioritized(callout{}, 100)))
	// Ahead of the default renderer (1000), so it replaces its blockquote.
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(callout{}, 100)))
}

// Transform runs as each paragraph closes, before inline parsing, so the
// marker line can be cut out and the title still parsed as markdown.
func (callout) Transform(p *ast.Paragraph, reader text.Reader, pc parser.Context) {
	quote, ok := p.Parent().(*ast.Blockquote)
	if !ok || quote.FirstChild() != p || p.Lines().Len() == 0 {
		return
	}
	types, _ := pc.Get(calloutKey).(map[string]Callout)
	src := reader.Source()
	first := p.Lines().At(0)
	m := calloutMarker.FindSubmatchIndex(first.Value(src))
	if m == nil {
		return
	}
	name := strings.ToLower(string(first.Value(src)[m[2]:m[3]]))
	c, ok := types[name]
	if !ok {
		return
	}

	title := &calloutTitle{Icon: c.Icon, Label: c.Title}
	if title.Label == "" {
		title.Label = defaultCalloutTitle(name)
	}
	if m[4] < m[5] {
		title.Lines().Append(text.NewSegment(first.Start+m[4], first.Start+m[5]))
	}
	quote.SetAttribute(calloutAttr, []byte(name))
	quote.InsertBefore(quote, p, title)

	lines := p.Lines()
	if lines.Len() == 1 {
		quote.RemoveChild(quote, p)
		return
	}
	rest := text.NewSegments()
	for i := 1; i < lines.Len(); i++ {
		rest.Append(lines.At(i))
	}
	p.SetLines(rest)
}

func defaultCalloutTitle(name string) string {
	t := strings.ReplaceAll(name, "-", " ")
	return strings.ToUpper(t[:1]) + t[1:]
}

func (callout) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindBlockquote, func(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		name, ok := n.AttributeString(string(calloutAttr))
		switch {
		case !ok && entering:
			w.WriteString("<blockquote>\n")
		case !ok:
			w.WriteString("</blockquote>\n")
		case entering:
			w.WriteString(`<div class="callout callout-`)
			w.Write(util.EscapeHTML(name.([]byte)))
			w.WriteString("\">\n")
		default:
			w.WriteString("</div>\n")
		}
		return ast.WalkContinue, nil
	})
	reg.Register(kindCalloutTitle, func(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		t := n.(*calloutTitle)
		if entering {
			w.WriteString(`<p class="callout-title"><span class="callout-icon">`)
			w.Write(util.EscapeHTML([]byte(t.Icon)))
			w.WriteString("</span>")
			if !t.HasChildren() {
				w.Write(util.EscapeHTML([]byte(t.Label)))
			}
		} else {
			w.WriteString("</p>\n")
		}
		return ast.WalkContinue, nil
	})
}

// CalloutTypes merges config's icons and titles over the built-in types.
// Empty fields keep the built-in values.
func CalloutTypes(custom map[string]Callout) map[string]Callout {
	all := make(map[string]Callout, len(BuiltinCallouts)+len(custom))
	for k, v := range BuiltinCallouts {
		all[k] = v
	}
	for k, v := range custom {
		b := all[k]
		if v.Icon != "" {
			b.Icon = v.Icon
		}
		if v.Title != "" {
			b.Title = v.Title
		}
		if b.Icon == "" {
			b.Icon = "●"
		}
		all[k] = b
	}
	return all
}
