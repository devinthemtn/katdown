import '@fontsource-variable/inter';
import '@fontsource-variable/jetbrains-mono';
import './app.css';

import {EditFile, GetDocument, OpenExternal, OpenFile, OpenRelative} from '../wailsjs/go/main/App';
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

async function load() {
    const doc = await GetDocument();
    lastDoc = doc;
    applyTheme(doc);
    style.textContent = doc.css;

    // Keep the reading position on live reload, start at the top for a new file.
    const sameFile = doc.path === currentPath;
    const scroll = sameFile ? window.scrollY : 0;
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
    search.refresh(sameFile);
}

// Links: external ones go to the system browser, #anchors scroll in place,
// relative .md files open in katdown. Nothing navigates the webview away.
content.addEventListener('click', (e) => {
    const a = e.target.closest('a[href]');
    if (!a) return;
    e.preventDefault();

    const href = a.getAttribute('href');
    if (href.startsWith('#')) {
        const target = document.getElementById(decodeURIComponent(href.slice(1)));
        target?.scrollIntoView();
    } else if (/^(https?|mailto):/i.test(href)) {
        OpenExternal(href);
    } else if (!/^[a-z][a-z0-9+.-]*:/i.test(href) && /\.(md|markdown|mdown|mkd)(#.*)?$/i.test(href)) {
        OpenRelative(href).catch((err) => console.warn('katdown:', err));
    }
});

initKeys({
    openFile: () => OpenFile(),
    reload: load,
    edit: () => EditFile().catch((err) => {
        content.prepend(errorBanner(`Could not start editor: ${err}`));
        window.scrollTo(0, 0);
    }),
});

// With theme = "system", follow the OS as it switches.
scheme.addEventListener('change', () => {
    if (lastDoc?.theme !== 'system') return;
    applyTheme(lastDoc);
    redrawDiagrams();
});

EventsOn('doc:changed', load);
load();
