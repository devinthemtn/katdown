# katdown demo

A tour of everything katdown renders. Open it with `katdown demo.md`, press `t` for the table of contents, `]]` / `[[` to hop between headings, and `?` for the key help.

## Highlighting

With the default settings, anything in curly braces is highlighted: Dear {customer name}, your order {order id} ships on {ship date}.

It works inside other formatting too: **bold {placeholder}**, *italic {placeholder}*, ~~struck {placeholder}~~, and in a [link to {somewhere}](https://example.com).

Some things are never highlighted:

- inline code: `{not highlighted}`
- math: $\{x \mid x > 0\}$
- an escaped brace: \{skipped once} but {this one} still is
- a bare URL: https://example.com/{path}

Add your own patterns in `config.toml`, for example to flag TODO and FIXME:

```toml
[highlight]
patterns = ['\{[^{}]+\}', '\bTODO\b', '\bFIXME\b']
color = "#e0af68"
```

With that config, this line lights up: TODO write the release notes, FIXME the flaky test.

## Mermaid diagrams

Diagrams are drawn offline and take their colors from the stylesheet, so they follow the dark and light themes.

### Flowchart

```mermaid
flowchart LR
    A[demo.md saved] --> B{Changed?}
    B -- yes --> C[Render markdown]
    B -- no --> D[Do nothing]
    C --> E[Highlight patterns]
    E --> F[Draw diagrams and math]
    F --> G((Window updates))
```

### Sequence diagram

```mermaid
sequenceDiagram
    participant U as User
    participant K as katdown
    participant G as Go backend
    U->>K: katdown notes.md
    K->>G: read file
    G-->>K: rendered HTML
    K-->>U: show document
    loop on every save
        G->>K: file changed
        K-->>U: re-render
    end
    Note over U,K: Ctrl-r reloads config.toml and style.css
```

### State diagram

```mermaid
stateDiagram-v2
    [*] --> Reading
    Reading --> Searching: /
    Searching --> Reading: Esc
    Reading --> Contents: t
    Contents --> Reading: Enter
    Reading --> [*]: q
```

### Class diagram

```mermaid
classDiagram
    class Settings {
        +string Theme
        +Font Font
        +Highlight Highlight
    }
    class Font {
        +string Body
        +string Mono
        +int Size
    }
    class Highlight {
        +[]string Patterns
        +string Color
    }
    Settings --> Font
    Settings --> Highlight
```

### Gantt chart

```mermaid
gantt
    title Release plan
    dateFormat YYYY-MM-DD
    section Features
    Config file       :done,    c1, 2026-10-01, 10d
    Highlight options :done,    c2, after c1, 7d
    Table of contents :active,  c3, after c2, 5d
    section Release
    Windows installer :         r1, after c3, 4d
    Tag v1.0          :milestone, after r1, 0d
```

### Pie chart

```mermaid
pie title Where the time goes
    "Writing" : 55
    "Reading" : 30
    "Fixing typos" : 15
```

## Math

Inline math sits in the text: Euler's identity $e^{i\pi} + 1 = 0$, and the area of a circle is $A = \pi r^2$.

Display math goes on its own lines:

$$
\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}
$$

$$
\begin{aligned}
\nabla \cdot \mathbf{E} &= \frac{\rho}{\varepsilon_0} \\
\nabla \times \mathbf{B} &= \mu_0 \mathbf{J} + \mu_0 \varepsilon_0 \frac{\partial \mathbf{E}}{\partial t}
\end{aligned}
$$

Dollar amounts stay plain text, like pandoc: it costs $5 and $10. Write \$ for a literal dollar sign.

## Code

Syntax highlighting uses chroma; pick the style with `[code] dark` and `light` in the config.

```go
package main

import "fmt"

func main() {
	for i := range 3 {
		fmt.Printf("hello {%d}\n", i) // braces in code are not highlighted
	}
}
```

```python
from dataclasses import dataclass

@dataclass
class Note:
    title: str
    tags: list[str]

    def slug(self) -> str:
        return self.title.lower().replace(" ", "-")
```

```javascript
const debounce = (fn, ms) => {
    let t;
    return (...args) => {
        clearTimeout(t);
        t = setTimeout(() => fn(...args), ms);
    };
};
```

```bash
make build && ./build/bin/katdown demo.md
```

```diff
- theme = "dark"
+ theme = "system"
```

## Text formatting

Plain text with **bold**, *italic*, ***bold italic***, ~~strikethrough~~ and `inline code`.
A line ending in a backslash\
forces a line break.

> A blockquote.
>
> > And a nested one, with **formatting** and a {placeholder}.

---

## Lists

1. Ordered item
2. Another item
   - nested bullet
   - another nested bullet
     1. deeper still
3. Last item

- [x] Config file
- [x] Highlight options
- [ ] Write the demo document
- [ ] Ship it

## Tables

| Feature     | Syntax                 | Notes                      |
|:------------|:----------------------:|---------------------------:|
| Highlight   | `{text}`               | configurable patterns      |
| Mermaid     | ` ```mermaid `         | bundled, works offline     |
| Math        | `$…$`, `$$…$$`         | rendered with KaTeX        |
| Footnotes   | `[^1]`                 | collected at the end       |

## Links

- Web links open in the browser: [Mermaid docs](https://mermaid.js.org)
- Autolinks work too: https://katex.org
- Relative `.md` links open in katdown: [README](README.md)
- Headings get ids, so in-page links work: [back to Highlighting](#highlighting)

## Footnotes

katdown supports footnotes[^footnotes], and they can hold formatting[^fmt].

[^footnotes]: This one is collected at the bottom of the page.
[^fmt]: Including **bold**, `code` and a {placeholder}.

## Raw HTML

Raw HTML is escaped rather than rendered, because the page shares a webview with the Go bindings: <b>this is not bold</b>.
