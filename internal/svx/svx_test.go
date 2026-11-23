package svx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// project makes a SvelteKit-like folder with package.json, a doc under
// src/lib/content and an image under src/lib/assets, and returns the doc's
// path and the image's.
func project(t *testing.T) (doc, img string) {
	t.Helper()
	root := t.TempDir()
	write := func(rel string) string {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	write("package.json")
	img = write("src/lib/assets/pics/turkey.svg")
	write("src/lib/content/recipes/near.png")
	return filepath.Join(root, "src", "lib", "content", "recipes", "doc.svx"), img
}

func TestPrepare(t *testing.T) {
	doc, img := project(t)
	imgURL := filepath.ToSlash(img)
	near := filepath.ToSlash(filepath.Join(filepath.Dir(doc), "near.png"))

	src := "---\ntitle: x\n---\n\n<script>\n\timport turkey from '$lib/assets/pics/turkey.svg';\n\timport near from \"./near.png?url\";\n\timport gone from '$lib/assets/gone.svg';\n\timport Thing from 'some-package';\n</script>\n\n" +
		"<img src={turkey} alt=\"t\" class=\"w\" />\n\n<img src={ near }>\n\n<img src={gone}>\n\n<p title={other}>x</p>\n\n" +
		"```svelte\n<script>\nimport a from './near.png';\n</script>\n<img src={turkey}>\n```\n"
	got := string(Prepare([]byte(src), doc))

	for _, want := range []string{
		"---\ntitle: x\n---\n", // front matter kept
		`<img src="` + imgURL + `" alt="t" class="w" />`,
		`<img src="` + near + `">`,
		"<img src={gone}>",       // missing file: left alone
		"<p title={other}>x</p>", // not an import
		"```svelte\n<script>\nimport a from './near.png';\n</script>\n<img src={turkey}>\n```\n", // code untouched
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "$lib") || strings.Count(got, "<script>") != 1 {
		t.Errorf("top-level script not removed:\n%s", got)
	}
}

func TestPrepareOneLineScript(t *testing.T) {
	doc, img := project(t)
	got := string(Prepare([]byte("<script>import p from '$lib/assets/pics/turkey.svg'</script>\n\n![](x.png)\n<img src={p}>\n"), doc))
	want := "\n![](x.png)\n<img src=\"" + filepath.ToSlash(img) + "\">\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProjectRootPrefersSvelteConfig(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "app", "src")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"svelte.config.js", "app/package.json"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(f)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got, ok := projectRoot(sub); !ok || got != root {
		t.Errorf("projectRoot = %q, %v; want %q", got, ok, root)
	}
}

func TestIsSvx(t *testing.T) {
	if !IsSvx("/a/b.svx") || !IsSvx("B.SVX") || IsSvx("b.md") {
		t.Error("IsSvx wrong")
	}
}
