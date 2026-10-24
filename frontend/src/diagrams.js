// Renders ```mermaid blocks. Mermaid is bundled with the app and only loaded
// when a document actually contains a diagram. Its colors come from the
// stylesheet's CSS variables so diagrams match the page (and user CSS).

const scheme = window.matchMedia('(prefers-color-scheme: dark)');

let mermaidPromise;
let lastRoot;
let idCounter = 0;

function getMermaid() {
    mermaidPromise ??= import('mermaid').then(({default: mermaid}) => mermaid);
    return mermaidPromise;
}

function themeFromCSS() {
    const css = getComputedStyle(document.documentElement);
    const v = (name, fallback) => css.getPropertyValue(name).trim() || fallback;
    const dark = v('--diagram-dark', 'true') !== 'false';
    const bg = v('--bg', dark ? '#0d1017' : '#ffffff');
    const surface = v('--surface-2', dark ? '#1a202b' : '#eceef2');
    const fg = v('--fg-strong', dark ? '#eef2f7' : '#111418');
    const muted = v('--muted', '#7d8899');
    const accent = v('--accent', '#7aa2f7');
    const accent2 = v('--accent-2', '#bb9af7');
    const border = v('--border-strong', '#323b4a');
    return {
        theme: 'base',
        themeVariables: {
            darkMode: dark,
            background: bg,
            fontFamily: v('--font-body', 'sans-serif'),
            fontSize: '15px',
            primaryColor: surface,
            primaryTextColor: fg,
            primaryBorderColor: accent,
            secondaryColor: v('--surface', surface),
            secondaryBorderColor: accent2,
            secondaryTextColor: fg,
            tertiaryColor: bg,
            tertiaryBorderColor: border,
            tertiaryTextColor: fg,
            lineColor: muted,
            textColor: fg,
            mainBkg: surface,
            nodeBorder: accent,
            clusterBkg: v('--surface', surface),
            clusterBorder: border,
            edgeLabelBackground: v('--surface', surface),
            noteBkgColor: v('--surface', surface),
            noteBorderColor: accent2,
            noteTextColor: fg,
            actorBkg: surface,
            actorBorder: accent,
            actorTextColor: fg,
            actorLineColor: muted,
            signalColor: fg,
            signalTextColor: fg,
            labelBoxBkgColor: surface,
            labelBoxBorderColor: accent,
            labelTextColor: fg,
            loopTextColor: fg,
            activationBkgColor: v('--surface', surface),
            activationBorderColor: accent2,
        },
    };
}

export async function renderDiagrams(root) {
    lastRoot = root;
    const blocks = [...root.querySelectorAll('pre.mermaid')];
    if (blocks.length === 0) return;

    const mermaid = await getMermaid();
    // Re-read the palette every time: the stylesheet is reloaded with the document.
    mermaid.initialize({startOnLoad: false, securityLevel: 'strict', ...themeFromCSS()});

    for (const pre of blocks) {
        // Keep the source so the diagram can be redrawn when colors change.
        pre.dataset.src ??= pre.textContent;
        pre.removeAttribute('data-error');
        if (pre.nextElementSibling?.classList.contains('mermaid-error')) pre.nextElementSibling.remove();
        try {
            const {svg} = await mermaid.render(`katdown-mermaid-${idCounter++}`, pre.dataset.src);
            pre.innerHTML = svg;
        } catch (err) {
            pre.textContent = pre.dataset.src;
            pre.dataset.error = '';
            const note = document.createElement('p');
            note.className = 'mermaid-error';
            note.textContent = `Mermaid: ${err?.message ?? err}`;
            pre.after(note);
        }
    }
}

// A stylesheet that follows the OS theme changes its variables; redraw to match.
scheme.addEventListener('change', () => {
    if (mermaidPromise && lastRoot) renderDiagrams(lastRoot);
});
