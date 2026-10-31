# katdown

A small markdown viewer built with Go and [Wails](https://wails.io).

I like a lot of developers am used to VIM style keybindings and might even like using terminal based editors, but whomever you are I bet you already have a favorite editor so I made this to be just a really good Markdown viewer so you get the best of all worlds.

```
katdown [-html] [-makeconfig] [-makestyle] [file.md]
```

- `katdown notes.md` opens the file. It re-renders automatically when the file is saved.
- `katdown` with no file shows a file picker; cancelling it quits.
- `katdown -html notes.md` turns off HTML sanitizing, so scripts, `<style>`, inline styles and iframes in the document work too. Only use it on files you trust: script can reach katdown's file and editor functions.
- `katdown -makeconfig` creates `~/.config/katdown/config.toml` (the standard config folder on each OS) with every setting at its default, then exits. katdown also does this by itself on first start.
- `katdown -makestyle` creates `style.css` in the same folder containing the default stylesheet, then exits.

## If you use linux through a windows computer using WSL/WSL2 in a mixed environment I got you.

- 'katdown-win notes.md' opens the file with the windows installed version of katdown.
- `katdown-win` with no file shows the windows file picker.

Neither ever overwrites an existing file.

Supports GitHub-flavored markdown (tables, task lists, strikethrough, autolinks), footnotes, syntax highlighting, [Mermaid](https://mermaid.js.org) diagrams and math (`$…$` inline, `$$…$$` on its own lines for display, rendered with [KaTeX](https://katex.org)). Mermaid and KaTeX are bundled, so everything renders offline. Like pandoc, `$5 and $10` stays plain text; write `\$` for a literal dollar sign. YAML front matter (a `---` block at the very top) is shown as a collapsed table instead of as text. Raw HTML in documents is sanitized: common tags such as `<details>`, `<kbd>`, `<sub>`, `<img width=…>` and `<p align="center">` work, while scripts, styles, event handlers and `javascript:` links are removed (start katdown with `-html` to keep everything).

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

Useful selectors: `.markdown-body` (the document), `pre.chroma` (code; token classes come from [chroma](https://github.com/alecthomas/chroma)), `pre.mermaid`, `mark.katdown-match` / `mark.katdown-match.current` (search hits), `#katdown-cmdline` (the search bar), `#katdown-help` (the `?` overlay), `#katdown-toc` (the `t` contents overlay), `.math` / `.katex` (formulas), `.katdown-highlight` (highlighted words), `.katdown-frontmatter` (the front matter; `display: none` hides it).

## Keys

| Key | Action |
|---|---|
| `j` / `k` | scroll down / up |
| `d` / `u`, `Ctrl-d` / `Ctrl-u` | half page down / up |
| `Space` / `f` / `Ctrl-f`, `b` / `Ctrl-b` | page down / up |
| `gg` / `G` | top / bottom |
| `Ctrl-+` / `Ctrl--`, `Ctrl-0` | bigger / smaller text, reset (not saved) |
| `/` | search (Enter to run, Esc to cancel); smartcase like vim |
| `n` / `N` | next / previous match |
| `Esc` | clear search |
| `e` | edit the file in your editor (the view reloads on save) |
| `Ctrl-o` | open another file |
| `H` / `L` | back / forward between files (followed links, `Ctrl-o`) |
| `Ctrl-r` | reload |
| `t` | table of contents (`j`/`k` to move, `Enter` to jump) |
| `]]` / `[[` | next / previous heading |
| `?` | show / hide key help |
| `q` | quit |

Links: web links open in the system browser, relative `.md` links open in katdown (`other.md#heading` scrolls to the heading), and `H` goes back to where you were.

## Building

Needs Go 1.26+, Node/npm, the Wails v3 CLI and the GTK/WebKit dev packages (`wails3 doctor` lists them; on Ubuntu 24.04 that's `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`). Install the CLI at the version in `go.mod`, with the same `gtk3` tag the build uses (Wails drops GTK 3 in v3.1, so upgrading past v3.0.x means switching to GTK 4):

```
go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
```

`wails3 task` runs the tasks in `Taskfile.yml` (a standalone [Task](https://taskfile.dev) works too: `task build`):

```
wails3 task build                  # -> bin/katdown
wails3 task dev FILE=README.md     # live-reloading dev mode
wails3 task test
wails3 task install                # copies to ~/.local/bin + app launcher entry
wails3 task build-windows          # -> bin/katdown.exe (cross-compiled)
wails3 task install-windows        # from WSL: copies to %LOCALAPPDATA%\Programs\katdown + Start Menu shortcut, .md handler, ~/.local/bin/katdown-win
wails3 task build-win-installer    # -> bin/katdown-amd64-installer.exe for Windows users (needs makensis)
wails3 task --list                 # everything else
```

The Windows build needs no extra toolchain; the target machine needs the WebView2 runtime (preinstalled on Windows 10/11). The installer is the one to hand to Windows users: it installs per user without an admin prompt, adds Start Menu and desktop shortcuts and the `.md` Open with entry, and installs WebView2 if it is missing (that step needs internet). Uninstall from Settings → Apps. Windows won't let an app make itself the default for `.md`; after installing, right-click a `.md` file → Open with → katdown → Always.

`wails3 task install-windows` also installs `katdown-win` in `~/.local/bin`, so `katdown-win notes.md` in the WSL shell opens the file in the Windows build (file arguments are converted with `wslpath -w`). Pick another name with `wails3 task install-windows WIN_WRAPPER=katdown`.

The Linux build uses GTK 3 with webkit2gtk-4.1. To build against GTK 4 and webkitgtk-6.0 instead (`libgtk-4-dev`, `libwebkitgtk-6.0-dev`), run `wails3 task build GTK=4`. webkit2gtk-4.0 is no longer supported.

Product name, version and copyright for the Windows exe, installer and macOS bundle live in `build/config.yml`; after changing them run `wails3 task update-build-assets`.

If the window is blank under WSLg or with some GPU drivers, try `WEBKIT_DISABLE_DMABUF_RENDERER=1 katdown file.md`.
