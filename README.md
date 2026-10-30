# katdown

A small markdown viewer built with Go and [Wails](https://wails.io).

```
katdown [-makeconfig] [-makestyle] [file.md]
```

- `katdown notes.md` opens the file. It re-renders automatically when the file is saved.
- `katdown` with no file shows a file picker; cancelling it quits.
- `katdown -makeconfig` creates `~/.config/katdown/config.toml` (the standard config folder on each OS) with every setting at its default, then exits. katdown also does this by itself on first start.
- `katdown -makestyle` creates `style.css` in the same folder containing the default stylesheet, then exits.

Neither ever overwrites an existing file.

Supports GitHub-flavored markdown (tables, task lists, strikethrough, autolinks), footnotes, syntax highlighting, [Mermaid](https://mermaid.js.org) diagrams and math (`$…$` inline, `$$…$$` on its own lines for display, rendered with [KaTeX](https://katex.org)). Mermaid and KaTeX are bundled, so everything renders offline. Like pandoc, `$5 and $10` stays plain text; write `\$` for a literal dollar sign. Raw HTML in documents is escaped.

## Settings

Settings live in `config.toml` in the config folder (`katdown -h` prints its path). katdown creates it with the defaults on first start; it never overwrites it, so delete it to get a fresh copy. Every setting is optional; anything left out uses the default below. Press `Ctrl-r` to apply edits; window size applies on the next start. Mistakes are shown at the top of the window, and the default is used for that setting.

```toml
theme = "dark"                 # dark | light | system (follows the OS)

[font]
body = "Inter Variable"        # any installed font; bundled: "Inter Variable", "JetBrains Mono Variable"
mono = "JetBrains Mono Variable"
size = 16                      # px

[layout]
width = "46rem"                # max width of the text column

[code]
dark = "tokyonight-night"      # any chroma style: https://xyproto.github.io/splash/docs/
light = "github"

[window]
width = 960
height = 900

[highlight]
patterns = ['\{[^{}]+\}']    # regular expressions (Go syntax); [] turns highlighting off
color = ""                     # CSS color; empty uses the theme's --highlight

[editor]
command = ""                   # editor for the e key, e.g. "code --wait"; empty uses $VISUAL
```

By default, anything in `{curly braces}` is highlighted, which suits templates and placeholders. Add patterns to mark other words, e.g. `patterns = ['\{[^{}]+\}', '\bTODO\b']`, using single quotes so backslashes are kept. Code, math and URLs are never highlighted, and `\{x}` skips one occurrence.

`e` opens the file in `editor.command` if set, otherwise `$VISUAL`, otherwise the system default (`xdg-open` on Linux, Notepad on Windows and under WSL without `xdg-open`). `$EDITOR` is not used: it is normally a terminal editor, and katdown has no terminal to run it in.

## Styling

The default look uses the bundled Inter and JetBrains Mono fonts, with a dark and a light palette.

If `style.css` exists in the config folder, it **replaces** the built-in style. Run `-makestyle` to get a full copy of the default to edit; delete the file to go back to the default. The palette is a set of CSS variables at the top of the file (Mermaid diagrams use them too), with the light palette under `:root[data-theme="light"]`. The settings still apply on top of a custom `style.css`: they set `data-theme` on `<html>`, the `--font-body`, `--font-mono`, `--font-size` and `--content-width` variables, and the code colors (which a `style.css` can override).

Useful selectors: `.markdown-body` (the document), `pre.chroma` (code; token classes come from [chroma](https://github.com/alecthomas/chroma)), `pre.mermaid`, `mark.katdown-match` / `mark.katdown-match.current` (search hits), `#katdown-cmdline` (the search bar), `#katdown-help` (the `?` overlay), `#katdown-toc` (the `t` contents overlay), `.math` / `.katex` (formulas), `.katdown-highlight` (highlighted words).

## Keys

| Key | Action |
|---|---|
| `j` / `k` | scroll down / up |
| `d` / `u`, `Ctrl-d` / `Ctrl-u` | half page down / up |
| `Space` / `f` / `Ctrl-f`, `b` / `Ctrl-b` | page down / up |
| `gg` / `G` | top / bottom |
| `/` | search (Enter to run, Esc to cancel); smartcase like vim |
| `n` / `N` | next / previous match |
| `Esc` | clear search |
| `e` | edit the file in your editor (the view reloads on save) |
| `Ctrl-o` | open another file |
| `Ctrl-r` | reload |
| `t` | table of contents (`j`/`k` to move, `Enter` to jump) |
| `]]` / `[[` | next / previous heading |
| `?` | show / hide key help |
| `q` | quit |

Links: web links open in the system browser, relative `.md` links open in katdown.

## Building

Needs Go, Node/npm, the Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) and the GTK/WebKit dev packages (`wails doctor` lists them).

```
make build                 # -> build/bin/katdown
make dev FILE=README.md    # live-reloading dev mode
make test
make install               # copies to ~/.local/bin
make build-windows         # -> build/bin/katdown.exe (cross-compiled)
make install-windows       # from WSL: copies to %LOCALAPPDATA%\Programs\katdown + Start Menu shortcut, .md handler
make installer-windows     # NSIS installer (needs makensis)
```

The Windows build needs no extra toolchain; the target machine needs the WebView2 runtime (preinstalled on Windows 10/11). Windows won't let an app make itself the default for `.md`; after installing, right-click a `.md` file → Open with → katdown → Always.

On distros that still ship `webkit2gtk-4.0`, run `make build TAGS=`.

If the window is blank under WSLg or with some GPU drivers, try `WEBKIT_DISABLE_DMABUF_RENDERER=1 katdown file.md`.
# katdown
