package render

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// math claims $…$ and $$…$$ so markdown doesn't mangle the TeX (a_1 b_1 would
// become emphasis) and emits it escaped in .math elements for KaTeX in the
// frontend to draw.
//
//	$x$, $$x$$      inline: <span class="math inline">x</span>
//	$$ on its own   block:  <div class="math display">x</div>
//
// Like pandoc, an inline opening $ must be followed by a non-space and the
// closing $ preceded by a non-space and not followed by a digit, so "$5 and
// $10" stays text. \$ is a literal dollar sign.
type math struct{}

var kindMath = ast.NewNodeKind("Math")
var kindMathBlock = ast.NewNodeKind("MathBlock")

type mathNode struct {
	ast.BaseInline
	display bool
	tex     []byte
}

func (n *mathNode) Kind() ast.NodeKind { return kindMath }

func (n *mathNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"TeX": string(n.tex)}, nil)
}

type mathBlockNode struct {
	ast.BaseBlock
	closed bool // the closing $$ has been seen
}

func (n *mathBlockNode) Kind() ast.NodeKind { return kindMathBlock }
func (n *mathBlockNode) IsRaw() bool        { return true }

func (n *mathBlockNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

func (math) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(mathBlockParser{}, 750)),
		parser.WithInlineParsers(util.Prioritized(mathInlineParser{}, 150)),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(mathRenderer{}, 100)))
}

// ---- Inline ----

type mathInlineParser struct{}

func (mathInlineParser) Trigger() []byte { return []byte{'$'} }

func (mathInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	opener := 0
	for opener < len(line) && line[opener] == '$' {
		opener++
	}
	if opener > 2 {
		return nil
	}
	display := opener == 2
	if !display && (len(line) < 2 || util.IsSpace(line[1])) {
		return nil
	}

	l, pos := block.Position()
	block.Advance(opener)
	var tex bytes.Buffer
	for {
		line, _ := block.PeekLine()
		if line == nil {
			block.SetPosition(l, pos)
			return nil
		}
		for i := 0; i < len(line); i++ {
			switch line[i] {
			case '\\':
				i++ // skip the escaped character, e.g. \$
			case '$':
				if !closes(line, i, opener) {
					continue
				}
				tex.Write(line[:i])
				block.Advance(i + opener)
				if tex.Len() == 0 {
					block.SetPosition(l, pos)
					return nil
				}
				return &mathNode{display: display, tex: bytes.Clone(tex.Bytes())}
			}
		}
		tex.Write(line)
		block.AdvanceLine()
	}
}

// closes reports whether the $ at line[i] ends math opened with opener $s.
func closes(line []byte, i, opener int) bool {
	if opener == 2 {
		return i+1 < len(line) && line[i+1] == '$'
	}
	if i == 0 || util.IsSpace(line[i-1]) {
		return false
	}
	return i+1 >= len(line) || line[i+1] < '0' || line[i+1] > '9'
}

// ---- Block ----

type mathBlockParser struct{}

func (mathBlockParser) Trigger() []byte { return []byte{'$'} }

func (mathBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || !bytes.HasPrefix(line[pos:], []byte("$$")) {
		return nil, parser.NoChildren
	}
	node := &mathBlockNode{}
	start := pos + 2
	rest := util.TrimRightSpace(line[start:])
	if end := bytes.Index(rest, []byte("$$")); end >= 0 {
		// $$ x $$ on one line; anything after the closing $$ makes it inline.
		if end+2 != len(rest) {
			return nil, parser.NoChildren
		}
		node.closed = true
		rest = rest[:end]
	}
	if !util.IsBlank(rest) {
		node.Lines().Append(text.NewSegment(segment.Start+start, segment.Start+start+len(rest)))
	}
	reader.AdvanceToEOL()
	return node, parser.NoChildren
}

func (mathBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	n := node.(*mathBlockNode)
	if n.closed {
		return parser.Close
	}
	line, segment := reader.PeekLine()
	trimmed := util.TrimRightSpace(line)
	if bytes.HasSuffix(trimmed, []byte("$$")) {
		n.closed = true
		trimmed = trimmed[:len(trimmed)-2]
	}
	if !util.IsBlank(trimmed) {
		n.Lines().Append(text.NewSegment(segment.Start, segment.Start+len(trimmed)))
	}
	reader.AdvanceToEOL()
	if n.closed {
		return parser.Close
	}
	return parser.Continue | parser.NoChildren
}

func (mathBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (mathBlockParser) CanInterruptParagraph() bool { return true }

func (mathBlockParser) CanAcceptIndentedLine() bool { return false }

// ---- Rendering ----

type mathRenderer struct{}

func (mathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMath, renderMath)
	reg.Register(kindMathBlock, renderMathBlock)
}

func renderMath(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*mathNode)
	class := "math inline"
	if n.display {
		class = "math display"
	}
	w.WriteString(`<span class="` + class + `">`)
	w.Write(util.EscapeHTML(n.tex))
	w.WriteString("</span>")
	return ast.WalkSkipChildren, nil
}

func renderMathBlock(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	w.WriteString(`<div class="math display">`)
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		if i > 0 {
			w.WriteByte('\n')
		}
		seg := lines.At(i)
		w.Write(util.EscapeHTML(seg.Value(src)))
	}
	w.WriteString("</div>\n")
	return ast.WalkSkipChildren, nil
}
