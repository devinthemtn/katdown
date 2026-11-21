import {beforeAll, beforeEach, describe, expect, it, vi} from 'vitest';
import * as app from '../bindings/katdown/app';
import {Events} from '@wailsio/runtime';
import * as command from '../src/command';
import {flush, key} from './helpers';

vi.mock('../src/diagrams', () => ({renderDiagrams: vi.fn(), redrawDiagrams: vi.fn()}));
vi.mock('../src/math', () => ({renderMath: vi.fn()}));

// A fake backend: OpenPath switches files and fires doc:changed like the Go side.
const files = {
    '/d/a.md': '<h2 id="top">A</h2><a href="b.md#sec">b</a> <a href="#end">end</a> <a href="https://x.org">x</a> <a href="pic.png">pic</a><h2 id="end">End</h2>',
    '/d/b.md': '<h2 id="sec">Sec</h2><a href="a.md">a</a>',
    '/d/c.md': '<p>c</p>',
};
let current = '/d/a.md';
let docExtra = {};
let changed;

const content = document.getElementById('content');
const html = () => content.innerHTML;
const click = (sel) => content.querySelector(sel).dispatchEvent(new MouseEvent('click', {bubbles: true, cancelable: true}));
const cmdMsg = () => document.querySelector('#katdown-command .msg').textContent;

async function settle() {
    for (let i = 0; i < 5; i++) await flush();
}

async function runCommand(text) {
    command.open();
    const input = document.querySelector('#katdown-command input');
    input.value = text;
    key('Enter', {}, input);
    await settle();
}

beforeAll(async () => {
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {});
    app.GetDocument.mockImplementation(async () => ({
        path: current, name: current.split('/').pop(), html: files[current] ?? '',
        css: 'body{}', theme: 'dark', toc: 'overlay', codeCss: {dark: 'DARK', light: 'LIGHT'},
        error: '', configError: '', ...docExtra,
    }));
    app.ResolveRelative.mockImplementation(async (href) => {
        const p = '/d/' + href.split('#')[0].replace(/^\.\//, '');
        if (!(p in files)) throw new Error(`no such file: ${p}`);
        return p;
    });
    app.OpenPath.mockImplementation(async (p) => {
        if (!(p in files)) throw new Error(`no such file: ${p}`);
        current = p;
        changed();
    });
    await import('../src/main');
    changed = Events.On.mock.calls.find(([name]) => name === 'doc:changed')[1];
    await settle();
});

beforeEach(async () => {
    docExtra = {};
    current = '/d/a.md';
    changed();
    await settle();
});

describe('loading', () => {
    it('shows the document with its styles and theme', () => {
        expect(html()).toBe(files['/d/a.md']);
        expect(document.getElementById('doc-style').textContent).toBe('body{}');
        expect(document.getElementById('code-style').textContent).toBe('DARK');
        expect(document.documentElement.dataset.theme).toBe('dark');
    });

    it('re-renders on doc:changed', async () => {
        files['/d/a.md'] += '<p>new</p>';
        changed();
        await settle();
        expect(html()).toContain('<p>new</p>');
    });

    it('shows file and config errors as banners', async () => {
        docExtra = {error: 'permission denied', configError: 'config.toml: bad theme'};
        changed();
        await settle();
        const banners = [...content.querySelectorAll('.katdown-error')].map((p) => p.textContent);
        expect(banners).toEqual(['config.toml: bad theme', 'Could not open /d/a.md: permission denied']);
    });
});

describe('links', () => {
    it('opens web links in the browser', () => {
        click('a[href^="https"]');
        expect(app.OpenExternal).toHaveBeenCalledWith('https://x.org');
    });

    it('scrolls to #anchors in place', () => {
        const end = document.getElementById('end');
        const spy = vi.spyOn(end, 'scrollIntoView');
        click('a[href="#end"]');
        expect(spy).toHaveBeenCalled();
        expect(app.OpenPath).not.toHaveBeenCalled();
    });

    it('opens relative markdown links in katdown at their anchor', async () => {
        click('a[href="b.md#sec"]');
        await settle();
        expect(app.OpenPath).toHaveBeenCalledWith('/d/b.md');
        expect(html()).toBe(files['/d/b.md']);
    });

    it('ignores links to other files', async () => {
        click('a[href="pic.png"]');
        await settle();
        expect(app.OpenPath).not.toHaveBeenCalled();
        expect(app.OpenExternal).not.toHaveBeenCalled();
    });
});

describe('H / L history', () => {
    it('goes back and forward between files', async () => {
        click('a[href="b.md#sec"]');
        await settle();
        key('H');
        await settle();
        expect(current).toBe('/d/a.md');
        key('L');
        await settle();
        expect(current).toBe('/d/b.md');
        key('L'); // nothing further forward
        await settle();
        expect(current).toBe('/d/b.md');
    });

    it(':e records the file it leaves, so H returns', async () => {
        await runCommand('e c.md');
        expect(current).toBe('/d/c.md');
        await runCommand('back');
        expect(current).toBe('/d/a.md');
    });

    it(':e shows an error for a missing file', async () => {
        await runCommand('e gone.md');
        expect(cmdMsg()).toBe('no such file: /d/gone.md');
        expect(current).toBe('/d/a.md');
        command.close();
    });
});

describe(':set', () => {
    it('theme switches the colors', async () => {
        await runCommand('set theme=light');
        expect(document.documentElement.dataset.theme).toBe('light');
        expect(document.getElementById('code-style').textContent).toBe('LIGHT');
        await runCommand('set theme=dark');
        expect(document.documentElement.dataset.theme).toBe('dark');
    });

    it('rejects values that are not allowed', async () => {
        await runCommand('set theme=blue');
        expect(cmdMsg()).toBe('theme must be one of: dark, light, system');
        await runCommand('set nope=1');
        expect(cmdMsg()).toBe('Unknown option: nope');
        command.close();
    });

    it('width sets --content-width and survives a reload', async () => {
        await runCommand('set width=60rem');
        expect(document.documentElement.style.getPropertyValue('--content-width')).toBe('60rem');
        changed();
        await settle();
        expect(document.documentElement.style.getPropertyValue('--content-width')).toBe('60rem');
    });

    it('toc=sidebar shows the sidebar', async () => {
        await runCommand('set toc=sidebar');
        const toc = document.getElementById('katdown-toc');
        expect(toc.classList.contains('sidebar')).toBe(true);
        expect(toc.hidden).toBe(false);
        // Switching while it is shown keeps it shown, now as the overlay.
        await runCommand('set toc=overlay');
        expect(toc.classList.contains('sidebar')).toBe(false);
        expect(document.body.classList.contains('katdown-toc-sidebar')).toBe(false);
        key('Escape');
        expect(toc.hidden).toBe(true);
    });
});
