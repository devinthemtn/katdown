// The `t` table of contents overlay, and heading lookups for `]]` / `[[`.

const content = document.getElementById('content');
const SELECTOR = 'h1[id], h2[id], h3[id], h4[id]';
const GG_TIMEOUT = 500;

const overlay = document.createElement('div');
overlay.id = 'katdown-toc';
overlay.hidden = true;
overlay.setAttribute('role', 'dialog');
overlay.setAttribute('aria-label', 'Table of contents');

const panel = document.createElement('div');
panel.className = 'panel';
const title = document.createElement('h2');
title.textContent = 'Contents';
const list = document.createElement('ol');
const hint = document.createElement('p');
hint.className = 'hint';
hint.textContent = 'j/k to move, Enter to jump, t or Esc to close';
panel.append(title, list, hint);
overlay.append(panel);
overlay.addEventListener('mousedown', (e) => e.target === overlay && close());
document.body.append(overlay);

let headings = [];
let selected = -1;
let lastG = 0;

const headingsNow = () => [...content.querySelectorAll(SELECTOR)];

// Distance from the heading to where scrollIntoView puts it (its scroll-margin).
const offset = (h) => h.getBoundingClientRect().top - (parseFloat(getComputedStyle(h).scrollMarginTop) || 0);

// The heading whose section is at the top of the viewport, or -1 above the first.
function currentIndex(hs) {
    let i = -1;
    for (const [j, h] of hs.entries()) {
        if (offset(h) > 1) break;
        i = j;
    }
    return i;
}

/** Scrolls to the next (dir 1) or previous (dir -1) heading, like vim's ]] and [[. */
export function jump(dir) {
    const hs = headingsNow();
    const cur = currentIndex(hs);
    if (dir > 0) {
        hs[cur + 1]?.scrollIntoView();
    } else if (cur >= 0 && offset(hs[cur]) < -1) {
        hs[cur].scrollIntoView(); // back to the start of this section
    } else if (cur > 0) {
        hs[cur - 1].scrollIntoView();
    } else {
        window.scrollTo({top: 0, behavior: 'instant'});
    }
}

export const isOpen = () => !overlay.hidden;

export function open() {
    headings = headingsNow();
    list.replaceChildren();
    if (headings.length === 0) {
        const li = document.createElement('li');
        li.className = 'empty';
        li.textContent = 'No headings';
        list.append(li);
    }
    headings.forEach((h, i) => {
        const li = document.createElement('li');
        li.className = `level-${h.tagName[1]}`;
        li.textContent = h.textContent;
        li.addEventListener('click', () => go(i));
        list.append(li);
    });
    overlay.hidden = false;
    select(Math.max(currentIndex(headings), 0));
}

export function close() {
    overlay.hidden = true;
}

function select(i) {
    if (headings.length === 0) return;
    list.children[selected]?.classList.remove('selected');
    selected = Math.max(0, Math.min(i, headings.length - 1));
    const li = list.children[selected];
    li.classList.add('selected');
    li.scrollIntoView({block: 'nearest'});
}

function go(i) {
    close();
    headings[i]?.scrollIntoView();
}

/** Handles a keydown while the overlay is open. */
export function onKey(e) {
    e.preventDefault();
    if (e.key !== 'g') lastG = 0;
    switch (e.key) {
        case 'j':
        case 'ArrowDown': select(selected + 1); break;
        case 'k':
        case 'ArrowUp': select(selected - 1); break;
        case 'G':
        case 'End': select(headings.length - 1); break;
        case 'Home': select(0); break;
        case 'g': {
            const now = Date.now();
            if (now - lastG < GG_TIMEOUT) {
                select(0);
                lastG = 0;
            } else {
                lastG = now;
            }
            break;
        }
        case 'Enter': go(selected); break;
        case 't':
        case 'q':
        case 'Escape': close(); break;
    }
}
