import {beforeEach, describe, expect, it, vi} from 'vitest';
import {ListFiles, Quit, RecentFiles} from '../bindings/katdown/app';
import * as command from '../src/command';
import {flush, key} from './helpers';

const overlay = document.getElementById('katdown-command');
const input = overlay.querySelector('input');
const msg = overlay.querySelector('.msg');
const items = () => [...overlay.querySelectorAll('li')].map((li) => li.querySelector('.label').textContent);

const A = {
    openFile: vi.fn(),
    openPath: vi.fn(),
    reload: vi.fn(),
    edit: vi.fn(),
    editConfig: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    setOption: vi.fn(() => ''),
    options: () => ({
        theme: {values: ['dark', 'light', 'system'], current: 'dark'},
        toc: {values: ['overlay', 'sidebar'], current: 'overlay'},
        width: {values: null, current: '46rem'},
    }),
    getZoom: vi.fn(() => 1.2),
    setZoom: vi.fn(),
};
command.initCommands(A);

async function type(text) {
    input.value = text;
    input.dispatchEvent(new Event('input'));
    await flush();
}

// Opens the window, types text and presses Enter.
async function run(text) {
    command.open();
    await type(text);
    key('Enter', {}, input);
    await flush();
}

beforeEach(() => {
    command.close();
    localStorage.clear();
});

describe('running commands', () => {
    it('runs a command by name, alias or unique prefix', async () => {
        await run('reload');
        await run('r');
        await run('rel');
        expect(A.reload).toHaveBeenCalledTimes(3);
        expect(command.isOpen()).toBe(false);
    });

    it('ignores a leading colon and spaces', async () => {
        await run(':  back');
        expect(A.back).toHaveBeenCalled();
    });

    it('opens a file or the dialog', async () => {
        await run('e notes.md');
        expect(A.openPath).toHaveBeenCalledWith('notes.md');
        await run('open');
        expect(A.openFile).toHaveBeenCalled();
    });

    it('passes arguments containing spaces and =', async () => {
        await run('e my notes.md');
        expect(A.openPath).toHaveBeenCalledWith('my notes.md');
        await run('set width = calc(1rem + 2px)');
        expect(A.setOption).toHaveBeenCalledWith('width', 'calc(1rem + 2px)');
    });

    it('edits config.toml and style.css', async () => {
        await run('config');
        await run('style');
        expect(A.editConfig.mock.calls).toEqual([['config'], ['style']]);
    });

    it('quits', async () => {
        await run('q');
        expect(Quit).toHaveBeenCalled();
    });

    it('scrolls to a percentage', async () => {
        const spy = vi.spyOn(window, 'scrollTo').mockImplementation(() => {});
        await run('50%');
        expect(spy).toHaveBeenCalledWith({top: expect.any(Number), behavior: 'instant'});
        spy.mockRestore();
    });
});

describe('errors', () => {
    it.each([
        ['xyz', 'Not a command: xyz'],
        ['t', 'Ambiguous: t could be top, toc'],
        ['h', 'Usage: heading <text>'],
        ['top now', 'top takes no argument'],
        ['zoom big', 'Not a zoom factor: big'],
        ['set nope', 'Usage: set <option>=<value> (options: theme, toc, width)'],
    ])(':%s', async (text, want) => {
        await run(text);
        expect(command.isOpen()).toBe(true);
        expect(input.value).toBe(text);
        expect(msg.textContent).toBe(want);
        expect(msg.classList.contains('error')).toBe(true);
    });

    it('shows an error thrown by an action', async () => {
        A.openPath.mockRejectedValueOnce(new Error('no such file'));
        await run('e gone.md');
        expect(msg.textContent).toBe('no such file');
    });

    it('shows an error returned by setOption', async () => {
        A.setOption.mockReturnValueOnce('theme must be one of: dark, light, system');
        await run('set theme=blue');
        expect(msg.textContent).toBe('theme must be one of: dark, light, system');
    });
});

describe(':set and :zoom', () => {
    it('shows one or all values', async () => {
        await run('set theme');
        expect(msg.textContent).toBe('theme=dark');
        expect(msg.classList.contains('error')).toBe(false);
        await run('set all');
        expect(msg.textContent).toBe('theme=dark  toc=overlay  width=46rem');
    });

    it('sets a factor or a percentage, and resets with none', async () => {
        await run('zoom 1.5');
        await run('zoom 120%');
        await run('zoom');
        expect(A.setZoom.mock.calls).toEqual([[1.5], [1.2], [1]]);
        await run('zoom 0');
        expect(msg.textContent).toBe('Not a zoom factor: 0');
    });
});

describe('suggestions', () => {
    it('lists commands matching what is typed', async () => {
        command.open();
        await type('re');
        expect(items()).toEqual(['recent', 'reload']);
        await type('');
        expect(items()).toContain('quit');
    });

    it('Tab moves through the list and fills the input', async () => {
        command.open();
        await type('re');
        key('Tab', {}, input);
        expect(input.value).toBe('recent ');
        key('Tab', {}, input);
        expect(input.value).toBe('reload');
        key('Tab', {shiftKey: true}, input);
        expect(input.value).toBe('recent ');
    });

    it('Tab on the only choice that takes more fills it and lists what comes next', async () => {
        command.open();
        await type('se');
        key('Tab', {}, input);
        await flush();
        expect(input.value).toBe('set ');
        expect(items()).toEqual(['theme=', 'toc=', 'width=']);
        key('Tab', {}, input);
        key('Enter', {}, input);
        await flush();
        expect(input.value).toBe('set theme=');
        expect(items()).toEqual(['theme=dark', 'theme=light', 'theme=system']);
    });

    it('completes files from ListFiles', async () => {
        ListFiles.mockResolvedValueOnce(['docs/', 'notes.md']);
        command.open();
        await type('e no');
        expect(ListFiles).toHaveBeenCalledWith('no');
        expect(items()).toEqual(['docs/', 'notes.md']);
    });

    it('completes recent files by name', async () => {
        RecentFiles.mockResolvedValue(['/a/one.md', '/b/two.md']);
        command.open();
        await type('recent tw');
        expect(items()).toEqual(['two.md']);
        key('Tab', {}, input);
        expect(input.value).toBe('recent /b/two.md');
        RecentFiles.mockResolvedValue([]);
    });

    it('drops a completion that arrives after the text changed', async () => {
        let resolve;
        ListFiles.mockReturnValueOnce(new Promise((r) => { resolve = r; }));
        command.open();
        input.value = 'e slow';
        input.dispatchEvent(new Event('input'));
        await type('re');
        resolve(['slow.md']);
        await flush();
        expect(items()).toEqual(['recent', 'reload']);
    });
});

describe('history and closing', () => {
    it('Up recalls earlier commands starting with what was typed', async () => {
        await run('reload');
        await run('back');
        await run('reload');
        command.open();
        key('ArrowUp', {}, input);
        expect(input.value).toBe('reload');
        key('ArrowUp', {}, input);
        expect(input.value).toBe('back');
        key('ArrowDown', {}, input);
        expect(input.value).toBe('reload');
        key('ArrowDown', {}, input);
        expect(input.value).toBe('');

        await type('b');
        key('ArrowUp', {}, input);
        expect(input.value).toBe('back');
    });

    it('saves history to localStorage', async () => {
        await run('top');
        expect(JSON.parse(localStorage.getItem('katdown-command-history'))).toContain('top');
    });

    it.each(['Escape', 'Backspace'])('%s closes the window', (k) => {
        command.open();
        key(k, {}, input);
        expect(command.isOpen()).toBe(false);
    });

    it('Backspace with text deletes instead of closing', () => {
        command.open('e x');
        key('Backspace', {}, input);
        expect(command.isOpen()).toBe(true);
    });
});
