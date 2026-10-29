// Vim-style `/` search: plain substring, smartcase, n/N to move between hits.

const content = document.getElementById('content');
const bar = document.getElementById('katdown-cmdline');
const input = bar.querySelector('input');
const status = bar.querySelector('.status');

const MARK = 'katdown-match';

let lastQuery = '';
let matches = [];
let current = -1;

export const isOpen = () => !bar.hidden && document.activeElement === input;

export function open() {
    bar.hidden = false;
    input.value = '';
    setStatus('');
    input.focus();
}

function close() {
    input.blur();
    if (!lastQuery) bar.hidden = true;
}

input.addEventListener('keydown', (e) => {
    e.stopPropagation();
    if (e.key === 'Enter') {
        e.preventDefault();
        const q = input.value;
        if (q) {
            lastQuery = q;
            run(true);
        } else if (lastQuery) {
            run(true); // like vim, an empty / repeats the last search
        }
        close();
    } else if (e.key === 'Escape') {
        e.preventDefault();
        input.value = '';
        close();
        if (lastQuery) showPosition();
    } else if (e.key === 'Backspace' && input.value === '') {
        e.preventDefault();
        close();
    }
});

// Drop focus from the bar when the page is clicked so vim keys work again.
content.addEventListener('mousedown', () => isOpen() && close());

/** Re-highlight after the document is re-rendered. */
export function refresh(sameFile) {
    matches = [];
    current = -1;
    if (!sameFile) {
        lastQuery = '';
        bar.hidden = true;
        return;
    }
    if (lastQuery) run(false);
}

/** Esc in normal mode: clear highlights and hide the bar. */
export function clear() {
    unmark();
    lastQuery = '';
    matches = [];
    current = -1;
    bar.hidden = true;
}

export function next(dir) {
    if (!lastQuery) return;
    if (matches.length === 0) {
        run(true);
        return;
    }
    if (current < 0) {
        // After a reload nothing is selected yet: start from the viewport.
        select(firstVisible());
        showPosition();
        return;
    }
    let i = current + dir;
    let wrapped = '';
    if (i >= matches.length) {
        i = 0;
        wrapped = 'search hit BOTTOM, continuing at TOP';
    } else if (i < 0) {
        i = matches.length - 1;
        wrapped = 'search hit TOP, continuing at BOTTOM';
    }
    select(i);
    bar.hidden = false;
    showPosition(wrapped);
}

function run(jump) {
    unmark();
    matches = mark(lastQuery);
    bar.hidden = false;
    if (matches.length === 0) {
        current = -1;
        setStatus(`Pattern not found: ${lastQuery}`);
        return;
    }
    if (jump) {
        select(firstVisible());
    } else {
        current = -1;
    }
    showPosition();
}

// First match at or below the top of the viewport, like vim's cursor.
function firstVisible() {
    const i = matches.findIndex((m) => m.getBoundingClientRect().top >= 0);
    return i === -1 ? 0 : i;
}

function select(i) {
    matches[current]?.classList.remove('current');
    current = i;
    const m = matches[i];
    m.classList.add('current');
    m.scrollIntoView({block: 'center', behavior: 'instant'});
}

function showPosition(note = '') {
    const pos = current >= 0 ? `[${current + 1}/${matches.length}]` : `${matches.length} matches`;
    input.value = lastQuery;
    setStatus(note ? `${note}  ${pos}` : pos);
}

function setStatus(text) {
    status.textContent = text;
}

function unmark() {
    const marks = content.querySelectorAll(`mark.${MARK}`);
    const parents = new Set();
    for (const m of marks) {
        parents.add(m.parentNode);
        m.replaceWith(...m.childNodes);
    }
    parents.forEach((p) => p.normalize());
}

function mark(query) {
    const smartcase = query !== query.toLowerCase();
    const needle = smartcase ? query : query.toLowerCase();

    const walker = document.createTreeWalker(content, NodeFilter.SHOW_TEXT, {
        acceptNode(node) {
            const el = node.parentElement;
            if (!el || el.closest('pre.mermaid, .math, svg, script, style')) return NodeFilter.FILTER_REJECT;
            return node.data ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_SKIP;
        },
    });

    const nodes = [];
    while (walker.nextNode()) nodes.push(walker.currentNode);

    const found = [];
    for (let node of nodes) {
        const hay = smartcase ? node.data : node.data.toLowerCase();
        let offset = 0;
        let idx;
        while ((idx = hay.indexOf(needle, offset)) !== -1) {
            // Split off the match and wrap it; continue in the remainder.
            const hit = node.splitText(idx - offset);
            node = hit.splitText(needle.length);
            const m = document.createElement('mark');
            m.className = MARK;
            hit.replaceWith(m);
            m.append(hit);
            found.push(m);
            offset = idx + needle.length;
        }
    }
    return found;
}
