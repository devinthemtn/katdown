// Vim-style navigation.
import {Quit} from '../wailsjs/go/main/App';
import * as help from './help';
import * as search from './search';
import * as toc from './toc';

const LINE = 60;
const GG_TIMEOUT = 500;
const SEQUENCE_KEYS = ['g', '[', ']'];

const scrollBy = (dy) => window.scrollBy({top: dy, behavior: 'instant'});
const page = () => window.innerHeight - 2 * LINE;
const half = () => window.innerHeight / 2;

export function initKeys({openFile, reload, edit}) {
    // Two-key sequences (gg, ]], [[): true on the second press of `key`.
    let pending = '';
    let pendingAt = 0;
    const repeated = (key) => {
        const now = Date.now();
        const hit = pending === key && now - pendingAt < GG_TIMEOUT;
        pending = hit ? '' : key;
        pendingAt = now;
        return hit;
    };

    window.addEventListener('keydown', (e) => {
        if (search.isOpen()) return; // the search input handles its own keys
        if (e.isComposing) return;

        // While help is up, only the keys that close it do anything.
        if (help.isOpen()) {
            if (e.key === '?' || e.key === 'Escape' || e.key === 'q') help.close();
            e.preventDefault();
            return;
        }
        if (toc.isOpen()) {
            toc.onKey(e);
            return;
        }

        const ctrl = e.ctrlKey && !e.altKey && !e.metaKey;
        if (ctrl) {
            const actions = {
                d: () => scrollBy(half()),
                u: () => scrollBy(-half()),
                f: () => scrollBy(page()),
                b: () => scrollBy(-page()),
                o: openFile,
                r: reload,
            };
            const action = actions[e.key.toLowerCase()];
            if (action) {
                e.preventDefault();
                action();
            }
            return;
        }
        if (e.ctrlKey || e.altKey || e.metaKey) return;

        if (!SEQUENCE_KEYS.includes(e.key)) pending = '';
        switch (e.key) {
            case 'j': scrollBy(LINE); break;
            case 'k': scrollBy(-LINE); break;
            case 'd': scrollBy(half()); break;
            case 'u': scrollBy(-half()); break;
            case ' ':
            case 'f': scrollBy(e.shiftKey ? -page() : page()); break;
            case 'b': scrollBy(-page()); break;
            case 'G': window.scrollTo({top: document.body.scrollHeight, behavior: 'instant'}); break;
            case 'g': if (repeated('g')) window.scrollTo({top: 0, behavior: 'instant'}); break;
            case ']': if (repeated(']')) toc.jump(1); break;
            case '[': if (repeated('[')) toc.jump(-1); break;
            case 'e': edit(); break;
            case 't': toc.open(); break;
            case '/': search.open(); break;
            case '?': help.open(); break;
            case 'n': search.next(1); break;
            case 'N': search.next(-1); break;
            case 'Escape': search.clear(); break;
            case 'q': Quit(); break;
            default: return;
        }
        e.preventDefault();
    });
}
