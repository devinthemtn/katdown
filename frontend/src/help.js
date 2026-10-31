// The `?` key overlay listing katdown's keys.

const KEYS = [
    ['Move', [
        [['j', 'k'], 'scroll down / up'],
        [['d', 'u'], 'half page down / up'],
        [['Space', 'b'], 'page down / up'],
        [['gg', 'G'], 'top / bottom'],
    ]],
    ['View', [
        [['Ctrl-+', 'Ctrl--'], 'bigger / smaller text'],
        [['Ctrl-0'], 'reset text size'],
    ]],
    ['Headings', [
        [['t'], 'table of contents'],
        [[']]', '[['], 'next / previous heading'],
    ]],
    ['Search', [
        [['/'], 'search (Enter runs, Esc cancels)'],
        [['n', 'N'], 'next / previous match'],
        [['Esc'], 'clear search'],
    ]],
    ['File', [
        [['e'], 'edit in your editor'],
        [['Ctrl-o'], 'open another file'],
        [['H', 'L'], 'back / forward between files'],
        [['Ctrl-r'], 'reload'],
        [['?'], 'toggle this help'],
        [['q'], 'quit'],
    ]],
];

const overlay = document.createElement('div');
overlay.id = 'katdown-help';
overlay.hidden = true;
overlay.setAttribute('role', 'dialog');
overlay.setAttribute('aria-label', 'Keyboard shortcuts');

const panel = document.createElement('div');
panel.className = 'panel';
const title = document.createElement('h2');
title.textContent = 'Keys';
panel.append(title);

// One table for all groups so the description column lines up.
const table = document.createElement('table');
for (const [group, rows] of KEYS) {
    const th = document.createElement('th');
    th.colSpan = 2;
    th.textContent = group;
    table.insertRow().append(th);
    for (const [keys, desc] of rows) {
        const tr = table.insertRow();
        const k = tr.insertCell();
        keys.forEach((key, i) => {
            if (i) k.append(' ');
            const kbd = document.createElement('kbd');
            kbd.textContent = key;
            k.append(kbd);
        });
        tr.insertCell().textContent = desc;
    }
}
panel.append(table);

const hint = document.createElement('p');
hint.className = 'hint';
hint.textContent = 'Press ? or Esc to close';
panel.append(hint);

overlay.append(panel);
overlay.addEventListener('mousedown', (e) => e.target === overlay && close());
document.body.append(overlay);

export const isOpen = () => !overlay.hidden;
export const open = () => { overlay.hidden = false; };
export const close = () => { overlay.hidden = true; };
export const toggle = () => { overlay.hidden = !overlay.hidden; };
