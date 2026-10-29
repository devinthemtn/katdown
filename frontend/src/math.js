// Renders $…$ / $$…$$ math (.math elements from the Go renderer) with KaTeX.
// KaTeX is bundled with the app (Mermaid uses it too) and only loaded when a
// document actually contains math. Formulas take the text color, so they
// follow the theme.

let katexPromise;

function getKatex() {
    katexPromise ??= Promise.all([
        import('katex'),
        import('katex/dist/katex.min.css'),
    ]).then(([{default: katex}]) => katex);
    return katexPromise;
}

export async function renderMath(root) {
    const els = [...root.querySelectorAll('.math')];
    if (els.length === 0) return;

    const katex = await getKatex();
    const errorColor = getComputedStyle(document.documentElement).getPropertyValue('--error').trim() || '#f7768e';
    for (const el of els) {
        el.dataset.src ??= el.textContent;
        // throwOnError: false shows a bad formula in the error color instead of failing.
        katex.render(el.dataset.src, el, {
            displayMode: el.classList.contains('display'),
            throwOnError: false,
            errorColor,
        });
    }
}
