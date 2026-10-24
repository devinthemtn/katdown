// Command chromagen writes the syntax-highlighting part of katdown's default
// stylesheet from a chroma theme.
package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

const style = "tokyonight-night"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: chromagen <out.css>")
		os.Exit(2)
	}
	f := html.New(html.WithClasses(true))

	var out bytes.Buffer
	out.WriteString("/* ---- Syntax highlighting (chroma " + style + ") ---- */\n")
	if err := f.WriteCSS(&out, styles.Get(style)); err != nil {
		fail(err)
	}
	if err := os.WriteFile(os.Args[1], out.Bytes(), 0o644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "chromagen:", err)
	os.Exit(1)
}
