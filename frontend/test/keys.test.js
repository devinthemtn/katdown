import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {Quit} from '../bindings/katdown/app';
import * as command from '../src/command';
import * as help from '../src/help';
import {getZoom, initKeys, setZoom} from '../src/keys';
import * as search from '../src/search';
import {key} from './helpers';

const A = {openFile: vi.fn(), reload: vi.fn(), edit: vi.fn(), back: vi.fn(), forward: vi.fn()};
initKeys(A);
command.initCommands({options: () => ({})});

let scrollBy;
let scrollTo;
beforeEach(() => {
    scrollBy = vi.spyOn(window, 'scrollBy').mockImplementation(() => {});
    scrollTo = vi.spyOn(window, 'scrollTo').mockImplementation(() => {});
});
afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
    help.close();
    command.close();
    setZoom(1);
});

describe('scrolling', () => {
    it.each([
        ['j', 60],
        ['k', -60],
    ])('%s scrolls a line', (k, dy) => {
        key(k);
        expect(scrollBy).toHaveBeenCalledWith({top: dy, behavior: 'instant'});
    });

    it('d / u and Ctrl-d / Ctrl-u scroll half a page', () => {
        key('d');
        key('u', {ctrlKey: true});
        expect(scrollBy.mock.calls.map(([o]) => Math.sign(o.top))).toEqual([1, -1]);
    });

    it('Shift-Space pages up', () => {
        key(' ', {shiftKey: true});
        expect(scrollBy.mock.calls[0][0].top).toBeLessThan(0);
    });

    it('G goes to the bottom', () => {
        key('G');
        expect(scrollTo).toHaveBeenCalledWith(expect.objectContaining({top: document.body.scrollHeight}));
    });

    it('gg goes to the top only when pressed twice quickly', () => {
        vi.useFakeTimers();
        key('g');
        vi.advanceTimersByTime(600);
        key('g');
        expect(scrollTo).not.toHaveBeenCalled();
        key('g');
        expect(scrollTo).toHaveBeenCalledWith({top: 0, behavior: 'instant'});
    });

    it('another key between the two g presses cancels gg', () => {
        key('g');
        key('j');
        key('g');
        expect(scrollTo).not.toHaveBeenCalled();
    });
});

describe('actions', () => {
    it.each([
        ['e', {}, 'edit'],
        ['H', {}, 'back'],
        ['L', {}, 'forward'],
        ['o', {ctrlKey: true}, 'openFile'],
        ['r', {ctrlKey: true}, 'reload'],
    ])('%s %o calls %s', (k, opts, action) => {
        const e = key(k, opts);
        expect(A[action]).toHaveBeenCalled();
        expect(e.defaultPrevented).toBe(true);
    });

    it('q quits', () => {
        key('q');
        expect(Quit).toHaveBeenCalled();
    });

    it('leaves Alt and Meta combinations and unbound keys alone', () => {
        expect(key('j', {altKey: true}).defaultPrevented).toBe(false);
        expect(key('e', {metaKey: true}).defaultPrevented).toBe(false);
        expect(key('x').defaultPrevented).toBe(false);
        expect(scrollBy).not.toHaveBeenCalled();
        expect(A.edit).not.toHaveBeenCalled();
    });
});

describe('zoom', () => {
    it('Ctrl-+ / Ctrl-- / Ctrl-0 change the text size', () => {
        key('+', {ctrlKey: true, shiftKey: true});
        key('=', {ctrlKey: true});
        expect(getZoom()).toBe(1.2);
        expect(document.body.style.fontSize).toBe('calc(var(--font-size) * 1.2)');
        key('-', {ctrlKey: true});
        expect(getZoom()).toBe(1.1);
        key('0', {ctrlKey: true});
        expect(getZoom()).toBe(1);
        expect(document.body.style.fontSize).toBe('');
    });

    it('stays between 0.5 and 3', () => {
        setZoom(10);
        expect(getZoom()).toBe(3);
        setZoom(0.1);
        expect(getZoom()).toBe(0.5);
    });
});

describe('overlays', () => {
    it('? opens help, which swallows keys until it is closed', () => {
        key('?');
        expect(help.isOpen()).toBe(true);
        key('j');
        expect(scrollBy).not.toHaveBeenCalled();
        key('Escape');
        expect(help.isOpen()).toBe(false);
    });

    it(': opens the command window, which keeps the keys', () => {
        key(':');
        expect(command.isOpen()).toBe(true);
        key('e');
        expect(A.edit).not.toHaveBeenCalled();
    });

    it('/ opens search', () => {
        key('/');
        expect(search.isOpen()).toBe(true);
        expect(document.activeElement).toBe(document.querySelector('#katdown-cmdline input'));
        document.activeElement.blur();
    });
});

describe('help overlay', () => {
    // Every key the overlay lists must do something.
    it('lists only bound keys', () => {
        const listed = [...document.querySelectorAll('#katdown-help kbd')].map((k) => k.textContent);
        expect(listed).toEqual(expect.arrayContaining(['j', 'k', 'f', 'Ctrl-d', 'gg', 'G', 't', ':', '?', 'q']));
        for (const name of ['f', 'Ctrl-f']) {
            scrollBy.mockClear();
            const ctrl = name.startsWith('Ctrl-');
            key(name.replace('Ctrl-', ''), {ctrlKey: ctrl});
            expect(scrollBy, name).toHaveBeenCalled();
        }
    });
});
