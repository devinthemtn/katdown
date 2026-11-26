# Changelog

## 2.7.0 — 2026-11-25

### Added

- Callouts: a blockquote starting with `[!NOTE]`, `[!TIP]`, `[!IMPORTANT]`, `[!WARNING]` or `[!CAUTION]` shows as a coloured box with an icon and title, like on GitHub. Text after the marker sets the title. Add your own types, or restyle these, with `[callouts.<type>]` in `config.toml` (`color`, `icon`, `title`).
- The Windows installer asks whether to add katdown to your `PATH`, so you can run `katdown notes.md` from a terminal. Uninstalling removes it again. Silent installs (`/S`) leave `PATH` alone.
- `task install-windows` (and `task install` in WSL) also checks whether katdown is on the Windows `PATH` and asks to add it. Answer up front with `WIN_PATH=true` or `WIN_PATH=false`.

## 2.1.2 — 2026-11-23

### Added

- `-scripts` flag: like `-html`, and also runs the document's `<script>` tags, for that session only. The window title shows "⚠ scripts enabled" while it's on.
- mdsvex (`.svx`) files: Svelte `<script>` blocks are hidden, and images imported in them (`import pic from '$lib/…'`, used as `src={pic}`) show. `.svx` files are listed in the file dialog and `:e` completion, and `.svx` links open in katdown.

## 2.1.0 — 2026-11-22

### Fixed

- Local images outside the document's folder now show, such as `![](../assets/pic.png)`, as do absolute paths, `~/` paths and `file://` URLs. This also applies to `<video>` and `<audio>`. Only image, video and audio files are served from outside the document's folder.

### Changed

- In WSL, `task install` also installs the Windows build (`install-windows`). Skip it with `task install WINDOWS=false`.

## 2.0.0 — 2026-11-21

### Upgrading from 1.x

- katdown now uses Wails v3 and builds with [Task](https://taskfile.dev) instead of `make`. Install the Wails CLI with `go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0`, then run `wails3 task install` (or `task install`). See [Install](README.md#install).
- Raw HTML in documents is no longer dropped: it is sanitized, so tags such as `<details>`, `<kbd>` and `<img width=…>` now show up. Scripts, styles and event handlers are still removed.
- Run `katdown -updateconfig` to add the new settings (`[html]`, `[toc]`) to your `config.toml` as comments.

### Added

- `:` command window with completion for commands, files, headings, recent files and settings: `:e`, `:recent`, `:h`, `:set`, `:zoom`, `:config`, `:style` and more. See [Commands](README.md#commands).
- Table of contents as a sidebar: `style = "sidebar"` under `[toc]`, or `:set toc=sidebar`.
- `H` / `L` to go back and forward between files, and `.md` links follow their `#anchor`.
- YAML front matter is shown as a collapsed table.
- `-html` flag and `unsafe` setting under `[html]` to keep raw HTML unsanitized, for files you trust.
- `-version` flag.
- `katDown` alias installed alongside `katdown`.
- Per-user Windows installer: `task build-win-installer`.
- MIT license.

### Changed

- The help overlay lists `f` and the `Ctrl-` scroll keys.
- mermaid 12.3, KaTeX 0.19.

### Development

- Frontend tests with Vitest; `task test` runs both the Go and frontend tests (`test:go`, `test:frontend`).
- More Go tests for serving local files, resolving links and loading documents.
- GitHub Actions runs the tests and builds on every push and pull request.

## 1.2.0 — 2026-10-30

- `Ctrl-+` / `Ctrl--` / `Ctrl-0` change the text size for the session.

## 1.1.0 — 2026-10-30

- App icon, Linux window icon and `.desktop` entry.
- `katdown-win` runs the Windows build from the WSL shell.
- An error is shown when the editor exits right away.

## 1.0.0 — 2026-10-29

- First release.
