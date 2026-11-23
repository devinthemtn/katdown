// Package svx makes mdsvex (.svx) files readable as plain markdown. Svelte
// <script> blocks are removed, and attributes such as src={pic} that use a
// file imported in them (import pic from '$lib/assets/pic.svg') get that
// file's path instead, so the image shows like any other local image. Other
// Svelte syntax (components, {#if} blocks, expressions) is left as it is.
package svx

import (
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// IsSvx reports whether path is an mdsvex file.
func IsSvx(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".svx")
}

var (
	scriptOpen  = regexp.MustCompile(`(?i)^\s*<script\b`)
	scriptClose = regexp.MustCompile(`(?i)</script\s*>`)
	fence       = regexp.MustCompile("^\\s{0,3}(`{3,}|~{3,})")
	// import name from 'path' (default imports only; a query such as ?url
	// is dropped).
	importDefault = regexp.MustCompile(`\bimport\s+([A-Za-z_$][\w$]*)\s+from\s+['"]([^'"?]+)(?:\?[^'"]*)?['"]`)
	// name={ident} or name={ ident } in a tag.
	attrExpr = regexp.MustCompile(`(\s[\w:-]+)=\{\s*([A-Za-z_$][\w$]*)\s*\}`)
)

// Prepare returns src, the contents of the .svx file at docPath, with its
// <script> blocks removed and imported files substituted into attributes.
// Text inside fenced code blocks is left alone.
func Prepare(src []byte, docPath string) []byte {
	dir := filepath.Dir(docPath)
	lines := strings.SplitAfter(string(src), "\n")

	// First pass: drop script blocks and collect their imports.
	imports := map[string]string{}
	var kept []string
	inScript := false
	var openFence string
	var script strings.Builder
	for _, line := range lines {
		switch {
		case inScript:
			script.WriteString(line)
			if scriptClose.MatchString(line) {
				inScript = false
				collectImports(script.String(), dir, imports)
				script.Reset()
			}
			continue
		case openFence != "":
			if m := fence.FindStringSubmatch(line); m != nil && m[1][0] == openFence[0] && len(m[1]) >= len(openFence) {
				openFence = ""
			}
		case fence.MatchString(line):
			openFence = fence.FindStringSubmatch(line)[1]
		case scriptOpen.MatchString(line):
			script.WriteString(line)
			if scriptClose.MatchString(line) {
				collectImports(script.String(), dir, imports)
				script.Reset()
			} else {
				inScript = true
			}
			continue
		}
		kept = append(kept, line)
	}
	if len(imports) == 0 {
		return []byte(strings.Join(kept, ""))
	}

	// Second pass: put the imported paths into attributes outside code.
	openFence = ""
	for i, line := range kept {
		if openFence != "" {
			if m := fence.FindStringSubmatch(line); m != nil && m[1][0] == openFence[0] && len(m[1]) >= len(openFence) {
				openFence = ""
			}
			continue
		}
		if m := fence.FindStringSubmatch(line); m != nil {
			openFence = m[1]
			continue
		}
		kept[i] = attrExpr.ReplaceAllStringFunc(line, func(s string) string {
			m := attrExpr.FindStringSubmatch(s)
			p, ok := imports[m[2]]
			if !ok {
				return s
			}
			return m[1] + `="` + html.EscapeString(p) + `"`
		})
	}
	return []byte(strings.Join(kept, ""))
}

// collectImports adds the default imports in script whose path katdown can
// find on disk to imports, as name → absolute path.
func collectImports(script, dir string, imports map[string]string) {
	for _, m := range importDefault.FindAllStringSubmatch(script, -1) {
		if p, ok := resolve(m[2], dir); ok {
			imports[m[1]] = filepath.ToSlash(p)
		}
	}
}

// resolve turns an import path into an absolute file path: $lib/ is the
// project's src/lib, and ./ or ../ paths are relative to dir. Package
// imports and files that don't exist are not resolved.
func resolve(spec, dir string) (string, bool) {
	var p string
	switch {
	case strings.HasPrefix(spec, "$lib/"):
		root, ok := projectRoot(dir)
		if !ok {
			return "", false
		}
		p = filepath.Join(root, "src", "lib", filepath.FromSlash(strings.TrimPrefix(spec, "$lib/")))
	case strings.HasPrefix(spec, "./"), strings.HasPrefix(spec, "../"):
		p = filepath.Join(dir, filepath.FromSlash(spec))
	default:
		return "", false
	}
	if fi, err := os.Stat(p); err != nil || fi.IsDir() {
		return "", false
	}
	return p, true
}

// projectRoot is the nearest folder at or above dir with a svelte.config.*
// file, or failing that, a package.json.
func projectRoot(dir string) (string, bool) {
	pkg := ""
	for d := dir; ; {
		if m, _ := filepath.Glob(filepath.Join(d, "svelte.config.*")); len(m) > 0 {
			return d, true
		}
		if pkg == "" {
			if _, err := os.Stat(filepath.Join(d, "package.json")); err == nil {
				pkg = d
			}
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	return pkg, pkg != ""
}
