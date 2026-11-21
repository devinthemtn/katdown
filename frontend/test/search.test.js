import {beforeEach, describe, expect, it} from 'vitest';
import * as search from '../src/search';
import {key} from './helpers';

const content = document.getElementById('content');
const bar = document.getElementById('katdown-cmdline');
const input = bar.querySelector('input');
const status = bar.querySelector('.status');
const marks = () => [...content.querySelectorAll('mark.katdown-match')];
const currentText = () => content.querySelector('mark.current')?.textContent;

function find(q) {
    search.open();
    input.value = q;
    key('Enter', {}, input);
}

beforeEach(() => {
    search.clear();
    content.innerHTML = `
        <p>Cat cat <b>CAT</b> concat</p>
        <pre class="mermaid">cat</pre>
        <span class="math">cat</span>
        <p>dog</p>`;
});

describe('search', () => {
    it('marks every match, case-insensitive for lowercase queries', () => {
        find('cat');
        expect(marks().map((m) => m.textContent)).toEqual(['Cat', 'cat', 'CAT', 'cat']);
        expect(status.textContent).toBe('[1/4]');
    });

    it('is case-sensitive once the query has a capital (smartcase)', () => {
        find('CAT');
        expect(marks().map((m) => m.textContent)).toEqual(['CAT']);
    });

    it('skips diagrams and math', () => {
        find('cat');
        expect(content.querySelector('pre.mermaid').innerHTML).toBe('cat');
        expect(content.querySelector('.math').innerHTML).toBe('cat');
    });

    it('n / N move through matches and wrap around', () => {
        find('cat');
        search.next(1);
        expect(status.textContent).toBe('[2/4]');
        search.next(-1);
        search.next(-1);
        expect(status.textContent).toBe('search hit TOP, continuing at BOTTOM  [4/4]');
        search.next(1);
        expect(status.textContent).toBe('search hit BOTTOM, continuing at TOP  [1/4]');
        expect(marks().filter((m) => m.classList.contains('current'))).toHaveLength(1);
    });

    it('reports no match', () => {
        find('zebra');
        expect(marks()).toHaveLength(0);
        expect(status.textContent).toBe('Pattern not found: zebra');
    });

    it('an empty search repeats the last one', () => {
        find('dog');
        search.clear();
        find('dog');
        find('');
        expect(marks().map((m) => m.textContent)).toEqual(['dog']);
    });

    it('clear removes the marks and restores the text', () => {
        const before = content.innerHTML;
        find('cat');
        search.clear();
        expect(content.innerHTML).toBe(before);
        expect(bar.hidden).toBe(true);
    });

    it('refresh re-marks the same file and forgets the search for a new one', () => {
        find('dog');
        content.innerHTML = '<p>dog dog</p>';
        search.refresh(true);
        expect(marks()).toHaveLength(2);
        expect(currentText()).toBeUndefined();
        expect(status.textContent).toBe('2 matches');

        search.refresh(false);
        search.next(1);
        expect(content.querySelector('mark.current')).toBeNull();
        expect(bar.hidden).toBe(true);
    });

    it('Escape cancels without searching', () => {
        search.open();
        input.value = 'cat';
        key('Escape', {}, input);
        expect(marks()).toHaveLength(0);
        expect(search.isOpen()).toBe(false);
    });
});
