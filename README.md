# katdown

A small markdown viewer built with Go and [Wails](https://wails.io).

```
katdown [-makeconfig] [file.md]
```

- `katdown notes.md` opens the file. It re-renders automatically when the file is saved.
- `katdown` with no file shows a file picker; cancelling it quits.
- `katdown -makeconfig` creates `~/.config/katdown/style.css` (the standard config folder on each OS) containing the default style, then exits. It never overwrites an existing file.

Supports GitHub-flavored markdown (tables, task lists, strikethrough, autolinks), footnotes, syntax highlighting and [Mermaid](https://mermaid.js.org) diagrams. Mermaid is bundled, so diagrams render offline. Raw HTML in documents is escaped.

## Styling

The default look is a dark theme using the bundled Inter and JetBrains Mono fonts, with tokyonight code colors.

If `style.css` exists in the config folder, it **replaces** the built-in style. Run `-makeconfig` to get a full copy of the default to edit; delete the file to go back to the default. The palette is a set of CSS variables at the top of the file (Mermaid diagrams use them too), and a light palette is included there, commented out.

Useful selectors: `.markdown-body` (the document), `pre.chroma` (code; token classes come from [chroma](https://github.com/alecthomas/chroma)), `pre.mermaid`, `mark.katdown-match` / `mark.katdown-match.current` (search hits), `#katdown-cmdline` (the search bar).

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
| `Ctrl-o` | open another file |
| `Ctrl-r` | reload |
| `q` | quit |

Links: web links open in the system browser, relative `.md` links open in katdown.

## Building

Needs Go, Node/npm, the Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) and the GTK/WebKit dev packages (`wails doctor` lists them).

```
make build                 # -> build/bin/katdown
make dev FILE=README.md    # live-reloading dev mode
make test
make install               # copies to ~/.local/bin
```

On distros that still ship `webkit2gtk-4.0`, run `make build TAGS=`.

If the window is blank under WSLg or with some GPU drivers, try `WEBKIT_DISABLE_DMABUF_RENDERER=1 katdown file.md`.
# katdown
