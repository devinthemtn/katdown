// Runs before each test file: builds the page from index.html and replaces
// the Go bindings and the Wails runtime with mocks.
import {readFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {beforeEach, vi} from 'vitest';

const html = readFileSync(resolve(process.cwd(), 'index.html'), 'utf8');
document.body.innerHTML = html.match(/<body>([\s\S]*)<\/body>/)[1].replace(/<script[\s\S]*?<\/script>/g, '');
for (const id of ['code-style', 'doc-style']) {
    const s = document.createElement('style');
    s.id = id;
    document.head.append(s);
}

vi.mock('../bindings/katdown/app', () => ({
    EditConfig: vi.fn(async () => {}),
    EditFile: vi.fn(async () => {}),
    GetDocument: vi.fn(async () => ({})),
    ListFiles: vi.fn(async () => []),
    OpenExternal: vi.fn(async () => {}),
    OpenFile: vi.fn(async () => {}),
    OpenPath: vi.fn(async () => {}),
    Quit: vi.fn(async () => {}),
    RecentFiles: vi.fn(async () => []),
    ResolveRelative: vi.fn(async (href) => href),
    Version: vi.fn(async () => 'test'),
}));

vi.mock('@wailsio/runtime', () => ({
    Events: {On: vi.fn()},
}));

// happy-dom has no layout; these only need to exist.
Element.prototype.scrollIntoView ??= function () {};

beforeEach(() => {
    vi.clearAllMocks();
});
