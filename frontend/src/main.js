import '@fontsource-variable/inter';
import '@fontsource-variable/jetbrains-mono';
import './app.css';

import {EditFile, GetDocument, OpenExternal, OpenFile, OpenPath, ResolveRelative} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';
import {redrawDiagrams, renderDiagrams} from './diagrams';
import {initKeys} from './keys';
import {renderMath} from './math';
import * as search from './search';

const style = document.getElementById('doc-style');
const codeStyle = document.getElementById('code-style');
const content = document.getElementById('content');
const scheme = window.matchMedia('(prefers-color-scheme: dark)');

let currentPath = '';
let lastDoc;

// Files left by following a link or opening another file, for H / L.
const back = [];
const forward = [];
// Where to put the view once pending.path loads: at pending.anchor, or at
// pending.scroll for a history move (which mustn't record itself again).
let pending = null;

// Sets data-theme on <html> from the settings (system follows the OS) and
// picks the matching code colors.
function applyTheme(doc) {
    const dark = doc.theme === 'system' ? scheme.matches : doc.theme !== 'light';
    document.documentElement.dataset.theme = dark ? 'dark' : 'light';
    codeStyle.textContent = dark ? doc.codeCss.dark : doc.codeCss.light;
}

function errorBanner(text) {
    const p = document.createElement('p');
    p.className = 'katdown-error';
    p.textContent = text;
    return p;
}

function showError(text) {
    content.prepend(errorBanner(text));
    window.scrollTo(0, 0);
}

function scrollToAnchor(id) {
    document.getElementById(decodeURIComponent(id))?.scrollIntoView();
}

async function load() {
    const doc = await GetDocument();
    lastDoc = doc;
    applyTheme(doc);
    style.textContent = doc.css;

    // Keep the reading position on live reload, start at the top for a new file.
    const sameFile = doc.path === currentPath;
    const target = pending?.path === doc.path ? pending : null;
    if (target) pending = null;
    if (!sameFile && currentPath && !target?.history) {
        back.push({path: currentPath, scroll: window.scrollY});
        forward.length = 0;
    }
    const scroll = sameFile ? window.scrollY : target?.scroll ?? 0;
    currentPath = doc.path;

    if (doc.error) {
        content.innerHTML = '';
        content.append(errorBanner(`Could not open ${doc.path}: ${doc.error}`));
    } else {
        content.innerHTML = doc.html;
    }
    if (doc.configError) content.prepend(errorBanner(doc.configError));

    await renderDiagrams(content);
    await renderMath(content);
    window.scrollTo(0, scroll);
    if (target?.anchor) scrollToAnchor(target.anchor);
    search.refresh(sameFile);
}

// H / L: reopen the file before or after this one, where it was left.
function move(from, to) {
    const entry = from.pop();
    if (!entry) return;
    to.push({path: currentPath, scroll: window.scrollY});
    pending = {...entry, history: true};
    OpenPath(entry.path).catch((err) => {
        to.pop();
        pending = null;
        showError(`Could not open ${entry.path}: ${err}`);
    });
}

// Links: external ones go to the system browser, #anchors scroll in place,
// relative .md files open in katdown. Nothing navigates the webview away.
content.addEventListener('click', (e) => {
    const a = e.target.closest('a[href]');
    if (!a) return;
    e.preventDefault();

    const href = a.getAttribute('href');
    if (href.startsWith('#')) {
        scrollToAnchor(href.slice(1));
    } else if (/^(https?|mailto):/i.test(href)) {
        OpenExternal(href);
    } else if (!/^[a-z][a-z0-9+.-]*:/i.test(href) && /\.(md|markdown|mdown|mkd)(#.*)?$/i.test(href)) {
        const anchor = href.includes('#') ? href.slice(href.indexOf('#') + 1) : '';
        ResolveRelative(href).then((path) => {
            if (path === currentPath) return scrollToAnchor(anchor);
            pending = {path, anchor};
            return OpenPath(path);
        }).catch((err) => console.warn('katdown:', err));
    }
});

initKeys({
    openFile: () => OpenFile(),
    reload: load,
    edit: () => EditFile().catch((err) => showError(`Could not start editor: ${err}`)),
    back: () => move(back, forward),
    forward: () => move(forward, back),
});

// With theme = "system", follow the OS as it switches.
scheme.addEventListener('change', () => {
    if (lastDoc?.theme !== 'system') return;
    applyTheme(lastDoc);
    redrawDiagrams();
});

EventsOn('doc:changed', load);
load();
