import '@fontsource-variable/inter';
import '@fontsource-variable/jetbrains-mono';
import './app.css';

import {GetDocument, OpenExternal, OpenFile, OpenRelative} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';
import {renderDiagrams} from './diagrams';
import {initKeys} from './keys';
import * as search from './search';

const style = document.getElementById('doc-style');
const content = document.getElementById('content');

let currentPath = '';

async function load() {
    const doc = await GetDocument();
    style.textContent = doc.css;

    // Keep the reading position on live reload, start at the top for a new file.
    const sameFile = doc.path === currentPath;
    const scroll = sameFile ? window.scrollY : 0;
    currentPath = doc.path;

    if (doc.error) {
        content.innerHTML = '';
        const p = document.createElement('p');
        p.className = 'katdown-error';
        p.textContent = `Could not open ${doc.path}: ${doc.error}`;
        content.append(p);
    } else {
        content.innerHTML = doc.html;
    }

    await renderDiagrams(content);
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
});

EventsOn('doc:changed', load);
load();
