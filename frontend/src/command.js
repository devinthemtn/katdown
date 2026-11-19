// The `:` command window: a box in the middle of the window for commands
// such as `:e notes.md`, `:h install` or `:set theme=light`.
//
// Typing filters a list of suggestions under the input. Tab / Shift-Tab (or
// Ctrl-n / Ctrl-p, or Down / Up once in the list) move through it and fill
// the input; Enter runs what is in the input, except that on a suggestion that
// needs more (a folder, `set`, `theme=`) it fills it in and lists what comes
// next. Up / Down before moving into the list recall earlier commands that
// start with what's typed, like vim.

import {ListFiles, Quit, RecentFiles} from '../bindings/katdown/app';
import * as help from './help';
import * as search from './search';
import * as toc from './toc';

const HISTORY_KEY = 'katdown-command-history';
const HISTORY_MAX = 100;

const overlay = document.createElement('div');
overlay.id = 'katdown-command';
overlay.hidden = true;
overlay.setAttribute('role', 'dialog');
overlay.setAttribute('aria-label', 'Command');

const panel = document.createElement('div');
panel.className = 'panel';
const line = document.createElement('div');
line.className = 'line';
const prompt = document.createElement('span');
prompt.className = 'prompt';
prompt.textContent = ':';
const input = document.createElement('input');
input.type = 'text';
input.spellcheck = false;
input.autocomplete = 'off';
input.setAttribute('aria-label', 'Command');
line.append(prompt, input);
const list = document.createElement('ul');
list.setAttribute('role', 'listbox');
const msg = document.createElement('p');
msg.className = 'msg';
panel.append(line, list, msg);
overlay.append(panel);
document.body.append(overlay);

// Clicking outside the panel closes it; clicking inside keeps the input focused.
overlay.addEventListener('mousedown', (e) => {
    if (e.target === overlay) close();
    else if (e.target !== input) e.preventDefault();
});

// Set by initCommands from main.js.
let A = {};

const scrollTo = (top) => window.scrollTo({top, behavior: 'instant'});
const maxScroll = () => document.documentElement.scrollHeight - window.innerHeight;

const COMMANDS = [
    {names: ['open', 'e', 'o'], args: '[file]', desc: 'open a file; with none, the file dialog',
        complete: completeFile, run: (arg) => (arg ? A.openPath(arg) : A.openFile())},
    {names: ['recent'], args: '[file]', desc: 'reopen a recently opened file',
        complete: completeRecent, run: runRecent},
    {names: ['heading', 'h'], args: '<text>', desc: 'jump to a heading',
        complete: completeHeading, run: runHeading},
    {names: ['top'], desc: 'go to the top', run: () => scrollTo(0)},
    {names: ['bottom'], desc: 'go to the bottom', run: () => scrollTo(maxScroll())},
    {names: ['toc'], desc: 'table of contents', run: () => toc.open()},
    {names: ['nohlsearch', 'noh'], desc: 'clear search highlights', run: () => search.clear()},
    {names: ['back'], desc: 'back to the previous file', run: () => A.back()},
    {names: ['forward'], desc: 'forward to the next file', run: () => A.forward()},
    {names: ['reload', 'r'], desc: 'reload the file and settings', run: () => A.reload()},
    {names: ['edit'], desc: 'edit the file in your editor', run: () => A.edit()},
    {names: ['set'], args: '<option>=<value>', desc: 'change a setting for this session',
        complete: completeSet, run: runSet},
    {names: ['zoom'], args: '[factor]', desc: 'text size, e.g. 1.2; with none, reset', run: runZoom},
    {names: ['config'], desc: 'edit config.toml (then :reload)', run: () => A.editConfig('config')},
    {names: ['style'], desc: 'edit style.css, creating it from the default', run: () => A.editConfig('style')},
    {names: ['help'], desc: 'key help', run: () => help.open()},
    {names: ['quit', 'q'], desc: 'quit katdown', run: () => Quit()},
];

/** Sets the actions the commands call; see main.js. */
export function initCommands(actions) {
    A = actions;
}

// ---- Parsing ----

// The command `name` refers to: an exact name or alias, else the only command
// with a name starting with it. Returns {cmd} or {error}.
function lookup(name) {
    const n = name.toLowerCase();
    const exact = COMMANDS.find((c) => c.names.includes(n));
    if (exact) return {cmd: exact};
    const hits = COMMANDS.filter((c) => c.names.some((x) => x.startsWith(n)));
    if (hits.length === 1) return {cmd: hits[0]};
    if (hits.length === 0) return {error: `Not a command: ${name}`};
    return {error: `Ambiguous: ${name} could be ${hits.map((c) => c.names[0]).join(', ')}`};
}

// Splits "e notes.md" into the command name and its argument.
function split(text) {
    const t = text.replace(/^[:\s]+/, '');
    const m = t.match(/^(\S*)(\s+)?(.*)$/s);
    return {name: m[1], spaced: Boolean(m[2]), arg: m[3].trim()};
}

// Runs a command line. Resolves to an error message, '' on success, or
// {info} for a message to show with the window left open.
async function execute(text, target) {
    const {name, arg} = split(text);
    if (!name) return '';
    const pct = name.match(/^(\d+(?:\.\d+)?)%$/);
    if (pct) {
        scrollTo(maxScroll() * Math.min(100, Number(pct[1])) / 100);
        return '';
    }
    const {cmd, error} = lookup(name);
    if (error) return error;
    if (cmd.args?.startsWith('<') && !arg) return `Usage: ${cmd.names[0]} ${cmd.args}`;
    if (!cmd.args && arg) return `${cmd.names[0]} takes no argument`;
    try {
        return (await cmd.run(arg, target)) ?? '';
    } catch (err) {
        return String(err?.message ?? err);
    }
}

// ---- Commands with arguments ----

const headingsMatching = (text) => {
    const t = text.toLowerCase();
    return toc.headingsNow().filter((h) => h.textContent.toLowerCase().includes(t));
};

function runHeading(arg, target) {
    const t = arg.toLowerCase();
    const hs = headingsMatching(arg);
    const h = target ?? hs.find((x) => x.textContent.trim().toLowerCase() === t) ?? hs[0];
    if (!h) return `No heading matching: ${arg}`;
    toc.goTo(h);
    return '';
}

function completeHeading(arg) {
    return headingsMatching(arg).map((h) => ({
        value: h.textContent.trim(),
        desc: h.tagName.toLowerCase(),
        level: Number(h.tagName[1]),
        target: h,
    }));
}

async function completeFile(arg) {
    const paths = await ListFiles(arg).catch(() => []);
    return (paths ?? []).map((p) => ({value: p, desc: p.endsWith('/') ? 'folder' : '', more: p.endsWith('/')}));
}

// Recent files whose path contains what was typed, shown by name and folder.
async function completeRecent(arg) {
    const t = arg.toLowerCase();
    const paths = await RecentFiles().catch(() => []);
    return (paths ?? [])
        .filter((p) => p.toLowerCase().includes(t))
        .map((p) => {
            const i = Math.max(p.lastIndexOf('/'), p.lastIndexOf('\\'));
            return {value: p, label: p.slice(i + 1), desc: p.slice(0, i)};
        });
}

async function runRecent(arg) {
    if (arg) return A.openPath(arg);
    const paths = await RecentFiles().catch(() => []);
    if (!paths?.length) return {info: 'No recent files'};
    return {input: 'recent '}; // reopen the window listing them
}

function runSet(arg) {
    const opts = A.options();
    if (!arg.includes('=')) {
        if (arg in opts) return {info: `${arg}=${opts[arg].current}`};
        if (arg === 'all') return {info: Object.entries(opts).map(([k, o]) => `${k}=${o.current}`).join('  ')};
        return `Usage: set <option>=<value> (options: ${Object.keys(opts).join(', ')})`;
    }
    const [name, ...rest] = arg.split('=');
    return A.setOption(name.trim(), rest.join('=').trim());
}

function completeSet(arg) {
    const opts = A.options();
    const eq = arg.indexOf('=');
    if (eq === -1) {
        return Object.entries(opts)
            .filter(([k]) => k.startsWith(arg))
            .map(([k, o]) => ({
                value: `${k}=`,
                desc: `now ${o.current}${o.values ? `; ${o.values.join(' | ')}` : ''}`,
                more: true,
            }));
    }
    const name = arg.slice(0, eq);
    const typed = arg.slice(eq + 1);
    const values = opts[name]?.values ?? [];
    return values
        .filter((v) => v.startsWith(typed))
        .map((v) => ({value: `${name}=${v}`, desc: v === opts[name].current ? 'current' : ''}));
}

function runZoom(arg) {
    if (!arg) {
        A.setZoom(1);
        return '';
    }
    const z = Number(arg.replace(/%$/, '')) / (arg.endsWith('%') ? 100 : 1);
    if (!Number.isFinite(z) || z <= 0) return `Not a zoom factor: ${arg}`;
    A.setZoom(z);
    return {info: `zoom ${A.getZoom()}`};
}

// ---- Suggestions ----

let items = [];
let selected = -1;
let seq = 0;

// Rebuilds the suggestion list for the input's current text.
async function suggest() {
    const n = ++seq;
    const text = input.value;
    const {name, spaced, arg} = split(text);
    let next = [];
    if (!spaced) {
        const typed = name.toLowerCase();
        next = COMMANDS
            .filter((c) => c.names.some((x) => x.startsWith(typed)))
            .map((c) => ({
                value: c.names[0] + (c.args ? ' ' : ''),
                label: c.names[0],
                aliases: c.names.slice(1),
                args: c.args,
                desc: c.desc,
                more: Boolean(c.complete),
            }));
    } else {
        const {cmd} = lookup(name);
        if (cmd?.complete) {
            const found = await cmd.complete(arg);
            if (n !== seq) return; // the text changed while this was loading
            next = found.map((it) => ({...it, label: it.label ?? it.value, value: `${name} ${it.value}`}));
        }
    }
    items = next;
    selected = -1;
    render();
}

function render() {
    list.replaceChildren(...items.map((it, i) => {
        const li = document.createElement('li');
        li.setAttribute('role', 'option');
        if (it.level) li.className = `level-${it.level}`;
        const label = document.createElement('span');
        label.className = 'label';
        label.textContent = it.label;
        li.append(label);
        if (it.aliases?.length) {
            const al = document.createElement('span');
            al.className = 'aliases';
            al.textContent = it.aliases.join(' ');
            li.append(al);
        }
        if (it.args) {
            const args = document.createElement('span');
            args.className = 'args';
            args.textContent = it.args;
            li.append(args);
        }
        if (it.desc) {
            const desc = document.createElement('span');
            desc.className = 'desc';
            desc.textContent = it.desc;
            li.append(desc);
        }
        li.addEventListener('click', () => {
            select(i);
            accept();
        });
        return li;
    }));
    list.hidden = items.length === 0;
}

function select(i) {
    if (items.length === 0) return;
    list.children[selected]?.classList.remove('selected');
    list.children[selected]?.setAttribute('aria-selected', 'false');
    selected = (i + items.length) % items.length;
    const li = list.children[selected];
    li.classList.add('selected');
    li.setAttribute('aria-selected', 'true');
    li.scrollIntoView({block: 'nearest'});
    input.value = items[selected].value;
}

// Enter: fill in a suggestion that needs more, otherwise run the input.
function accept() {
    const it = items[selected];
    if (it?.more) {
        input.value = it.value;
        setMessage('');
        suggest();
        return;
    }
    run(input.value, it?.target);
}

async function run(text, target) {
    const t = text.trim();
    if (t) remember(t);
    close();
    const result = await execute(t, target);
    if (typeof result?.input === 'string') open(result.input);
    else if (typeof result?.info === 'string') open('', result.info);
    else if (typeof result === 'string' && result) open(t, result, true);
}

// ---- History ----

let history = [];
try {
    history = JSON.parse(localStorage.getItem(HISTORY_KEY)) ?? [];
    if (!Array.isArray(history)) history = [];
} catch {
    history = [];
}
// Position while browsing history with Up / Down, and what was typed before.
let histPos = -1;
let histPrefix = '';

function remember(text) {
    history = history.filter((h) => h !== text);
    history.push(text);
    if (history.length > HISTORY_MAX) history = history.slice(-HISTORY_MAX);
    try {
        localStorage.setItem(HISTORY_KEY, JSON.stringify(history));
    } catch {
        // Not saved; it still works for this session.
    }
}

// Up (dir -1) / Down (dir 1): the earlier / later command starting with what
// was typed before browsing began.
function browseHistory(dir) {
    if (histPos === -1) {
        histPrefix = input.value;
        histPos = history.length;
    }
    let i = histPos + dir;
    while (i >= 0 && i < history.length && !history[i].startsWith(histPrefix)) i += dir;
    if (i < 0) return;
    histPos = Math.min(i, history.length);
    input.value = histPos === history.length ? histPrefix : history[histPos];
    setMessage('');
    suggest();
}

// ---- Open / close ----

function setMessage(text, error = false) {
    msg.textContent = text;
    msg.classList.toggle('error', error);
    msg.hidden = !text;
}

export const isOpen = () => !overlay.hidden;

/** Shows the window with `text` in the input and an optional message. */
export function open(text = '', message = '', error = false) {
    overlay.hidden = false;
    input.value = text;
    histPos = -1;
    setMessage(message, error);
    suggest();
    input.focus();
}

export function close() {
    overlay.hidden = true;
    input.blur();
    seq++; // drop any completion still loading
}

input.addEventListener('input', () => {
    histPos = -1;
    setMessage('');
    suggest();
});

input.addEventListener('keydown', (e) => {
    e.stopPropagation();
    if (e.isComposing) return;
    const ctrl = e.ctrlKey && !e.altKey && !e.metaKey;
    const key = ctrl ? `C-${e.key.toLowerCase()}` : e.key;
    switch (key) {
        case 'Enter': accept(); break;
        case 'Escape':
        case 'C-c':
        case 'C-[': close(); break;
        case 'Backspace':
            if (input.value !== '') return;
            close();
            break;
        case 'Tab':
            if (items.length === 1 && items[0].more) {
                selected = 0;
                accept(); // the only choice: fill it in and list what comes next
            } else {
                select(selected === -1 ? (e.shiftKey ? -1 : 0) : selected + (e.shiftKey ? -1 : 1));
            }
            break;
        case 'C-n': select(selected + 1); break;
        case 'C-p': select(selected - 1); break;
        case 'ArrowDown':
            if (selected === -1 && histPos !== -1) browseHistory(1);
            else select(selected + 1);
            break;
        case 'ArrowUp':
            if (selected === -1) browseHistory(-1);
            else select(selected - 1);
            break;
        default: return;
    }
    e.preventDefault();
});
