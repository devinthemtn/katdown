package render

import (
	"regexp"
	"slices"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// highlight wraps document text matching the configured patterns in
// <span class="katdown-highlight">. Code, math, autolinks, image alt text and
// raw HTML are left alone, and a match preceded by a backslash (\{x}) is
// not highlighted.
type highlight struct{}

// highlightKey holds the []*regexp.Regexp for one Render call.
var highlightKey = parser.NewContextKey()

var kindHighlight = ast.NewNodeKind("Highlight")

type highlightNode struct {
	ast.BaseInline
}

func (n *highlightNode) Kind() ast.NodeKind { return kindHighlight }

func (n *highlightNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

func (highlight) Extend(m goldmark.Markdown) {
	// After the other transformers (footnotes use 999), so the tree is final.
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(highlight{}, 2000)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(highlight{}, 100)))
}

func (highlight) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	patterns, _ := pc.Get(highlightKey).([]*regexp.Regexp)
	if len(patterns) == 0 {
		return
	}
	var texts []*ast.Text
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindCodeSpan, ast.KindCodeBlock, ast.KindFencedCodeBlock,
			ast.KindHTMLBlock, ast.KindRawHTML, ast.KindAutoLink, ast.KindImage:
			return ast.WalkSkipChildren, nil
		case ast.KindText:
			texts = append(texts, n.(*ast.Text))
		}
		return ast.WalkContinue, nil
	})
	// goldmark splits text at characters like _ < &, so match over runs of
	// adjacent text nodes on the same line.
	src := reader.Source()
	for len(texts) > 0 {
		n := 1
		for n < len(texts) && texts[n].PreviousSibling() == texts[n-1] &&
			!texts[n-1].SoftLineBreak() && !texts[n-1].HardLineBreak() {
			n++
		}
		highlightRun(texts[:n], src, patterns)
		texts = texts[n:]
	}
}

// highlightRun replaces adjacent text nodes with plain and highlighted pieces.
func highlightRun(run []*ast.Text, src []byte, patterns []*regexp.Regexp) {
	var value []byte
	starts := make([]int, len(run)) // offset of each node in value
	for i, t := range run {
		starts[i] = len(value)
		value = append(value, t.Segment.Value(src)...)
	}
	// srcPos maps an offset in value to its position in src.
	srcPos := func(off int) int {
		i := len(run) - 1
		for starts[i] > off {
			i--
		}
		return run[i].Segment.Start + off - starts[i]
	}

	// Match ranges from every pattern, merged where they overlap.
	var ranges [][]int
	for _, re := range patterns {
		for _, m := range re.FindAllIndex(value, -1) {
			p := srcPos(m[0])
			if m[0] < m[1] && (p == 0 || src[p-1] != '\\') {
				ranges = append(ranges, m)
			}
		}
	}
	if len(ranges) == 0 {
		return
	}
	slices.SortFunc(ranges, func(a, b []int) int { return a[0] - b[0] })
	merged := ranges[:1]
	for _, r := range ranges[1:] {
		last := merged[len(merged)-1]
		if r[0] <= last[1] {
			last[1] = max(last[1], r[1])
		} else {
			merged = append(merged, r)
		}
	}

	parent := run[0].Parent()
	var h *highlightNode // the span being filled, if any
	m := 0               // next range in merged
	for i, t := range run {
		lo, hi := starts[i], starts[i]+t.Segment.Len()
		for pos := lo; pos < hi; {
			for m < len(merged) && merged[m][1] <= pos {
				m++
			}
			inside := m < len(merged) && merged[m][0] <= pos
			stop := hi
			if m < len(merged) {
				if inside {
					stop = min(hi, merged[m][1])
				} else {
					stop = min(hi, merged[m][0])
				}
			}
			piece := newPiece(t, pos-lo, stop-lo)
			if inside {
				if h == nil {
					h = &highlightNode{}
					parent.InsertBefore(parent, t, h)
				}
				h.AppendChild(h, piece)
				if stop == merged[m][1] {
					h = nil
				}
			} else {
				parent.InsertBefore(parent, t, piece)
			}
			pos = stop
		}
		// The node's line break goes after everything cut from it, outside any span.
		if t.SoftLineBreak() || t.HardLineBreak() {
			br := newPiece(t, hi-lo, hi-lo)
			br.SetSoftLineBreak(t.SoftLineBreak())
			br.SetHardLineBreak(t.HardLineBreak())
			parent.InsertBefore(parent, t, br)
		}
		parent.RemoveChild(parent, t)
	}
}

// newPiece returns the part of t from start to stop (offsets in its segment).
func newPiece(t *ast.Text, start, stop int) *ast.Text {
	p := ast.NewTextSegment(text.NewSegment(t.Segment.Start+start, t.Segment.Start+stop))
	p.SetRaw(t.IsRaw())
	return p
}

func (highlight) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindHighlight, func(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			w.WriteString(`<span class="katdown-highlight">`)
		} else {
			w.WriteString("</span>")
		}
		return ast.WalkContinue, nil
	})
}
