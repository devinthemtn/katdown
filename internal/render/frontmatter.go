package render

import (
	"bytes"
	"html"
	"strings"

	"go.yaml.in/yaml/v3"
)

// splitFrontMatter separates a YAML block between --- lines at the very start
// of src from the markdown after it. ok is false when there is none; a ---
// with no closing line is an ordinary thematic break.
func splitFrontMatter(src []byte) (front, rest []byte, ok bool) {
	line, next := cutLine(src)
	if string(line) != "---" {
		return nil, src, false
	}
	for start := next; next < len(src); {
		line, end := cutLine(src[next:])
		if s := string(line); s == "---" || s == "..." {
			return src[start:next], src[next+end:], true
		}
		next += end
	}
	return nil, src, false
}

// cutLine returns the first line of b without its line ending or trailing
// blanks, and the length of the line including its ending.
func cutLine(b []byte) (line []byte, n int) {
	n = len(b)
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		n = i + 1
	}
	return bytes.TrimRight(b[:n], " \t\r\n"), n
}

// frontMatterHTML shows front matter as a collapsed table of its keys. Front
// matter that isn't a YAML mapping is shown as it is, with the reason.
func frontMatterHTML(src []byte) string {
	var doc yaml.Node
	err := yaml.Unmarshal(src, &doc)
	if err == nil && len(doc.Content) == 0 {
		return "" // empty
	}
	var b strings.Builder
	b.WriteString(`<details class="katdown-frontmatter">`)
	if err != nil || doc.Content[0].Kind != yaml.MappingNode {
		reason := "not a list of keys"
		if err != nil {
			reason = err.Error()
		}
		b.WriteString("<summary>Front matter (" + html.EscapeString(reason) + ")</summary>\n<pre><code>")
		b.WriteString(html.EscapeString(string(src)))
		b.WriteString("</code></pre></details>\n")
		return b.String()
	}
	b.WriteString("<summary>Front matter</summary>\n<table><tbody>\n")
	m := doc.Content[0].Content
	for i := 0; i+1 < len(m); i += 2 {
		b.WriteString("<tr><th>" + html.EscapeString(m[i].Value) + "</th><td>")
		b.WriteString(yamlValueHTML(m[i+1]))
		b.WriteString("</td></tr>\n")
	}
	b.WriteString("</tbody></table></details>\n")
	return b.String()
}

// yamlValueHTML shows a scalar as text, a list of scalars comma-separated,
// and anything else as YAML.
func yamlValueHTML(n *yaml.Node) string {
	switch n.Kind {
	case yaml.ScalarNode:
		return html.EscapeString(n.Value)
	case yaml.SequenceNode:
		items := make([]string, 0, len(n.Content))
		for _, c := range n.Content {
			if c.Kind != yaml.ScalarNode {
				items = nil
				break
			}
			items = append(items, html.EscapeString(c.Value))
		}
		if items != nil {
			return strings.Join(items, ", ")
		}
	}
	out, err := yaml.Marshal(n)
	if err != nil {
		return ""
	}
	return "<pre><code>" + html.EscapeString(strings.TrimRight(string(out), "\n")) + "</code></pre>"
}
