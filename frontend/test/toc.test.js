import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import * as toc from '../src/toc';

const content = document.getElementById('content');
const overlay = document.getElementById('katdown-toc');
const entries = () => [...overlay.querySelectorAll('li')];
const selected = () => overlay.querySelector('li.selected')?.textContent;
const press = (k) => toc.onKey(new KeyboardEvent('keydown', {key: k, cancelable: true}));

beforeEach(() => {
    content.innerHTML = '<h1 id="a">Alpha</h1><h2 id="b">Beta</h2><h3 id="c">Gamma</h3><h5 id="x">skipped</h5><h2>no id</h2>';
    // happy-dom has no layout: put the headings below the top of the viewport.
    [...content.children].forEach((h, i) => {
        h.getBoundingClientRect = () => ({top: 100 * (i + 1)});
    });
});
afterEach(() => {
    toc.close();
    toc.setMode('overlay');
});

describe('table of contents', () => {
    it('lists h1-h4 headings that have ids, by level', () => {
        toc.open();
        expect(entries().map((li) => [li.textContent, li.className.replace(' selected', '')]))
            .toEqual([['Alpha', 'level-1'], ['Beta', 'level-2'], ['Gamma', 'level-3']]);
        expect(toc.isOpen()).toBe(true);
    });

    it('says when there are no headings', () => {
        content.innerHTML = '<p>none</p>';
        toc.open();
        expect(entries().map((li) => li.textContent)).toEqual(['No headings']);
        press('j'); // nothing to select; must not throw
    });

    it('j / k / G / gg move the selection, clamped at the ends', () => {
        toc.open();
        expect(selected()).toBe('Alpha');
        press('k');
        expect(selected()).toBe('Alpha');
        press('j');
        press('j');
        press('j');
        expect(selected()).toBe('Gamma');
        press('g');
        press('g');
        expect(selected()).toBe('Alpha');
        press('G');
        expect(selected()).toBe('Gamma');
    });

    it('Enter jumps to the heading and closes the overlay', () => {
        const spy = vi.spyOn(document.getElementById('b'), 'scrollIntoView');
        toc.open();
        press('j');
        press('Enter');
        expect(spy).toHaveBeenCalled();
        expect(toc.isVisible()).toBe(false);
    });

    it('a sidebar stays visible after a jump but gives the keys back', () => {
        toc.setMode('sidebar');
        toc.open();
        expect(document.body.classList.contains('katdown-toc-sidebar')).toBe(true);
        press('Enter');
        expect(toc.isVisible()).toBe(true);
        expect(toc.isOpen()).toBe(false);
        press('Escape');
        expect(toc.isVisible()).toBe(false);
        expect(document.body.classList.contains('katdown-toc-sidebar')).toBe(false);
    });

    it('refresh rebuilds a shown list after the document changes', () => {
        toc.open();
        content.innerHTML = '<h2 id="z">Zeta</h2>';
        toc.refresh();
        expect(entries().map((li) => li.textContent)).toEqual(['Zeta']);
    });

    it('clicking outside the overlay closes it', () => {
        toc.open();
        overlay.dispatchEvent(new MouseEvent('mousedown', {bubbles: true}));
        expect(toc.isVisible()).toBe(false);
    });
});
