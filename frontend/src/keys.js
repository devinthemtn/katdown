// Vim-style navigation.
import {Quit} from '../wailsjs/go/main/App';
import * as help from './help';
import * as search from './search';

const LINE = 60;
const GG_TIMEOUT = 500;

const scrollBy = (dy) => window.scrollBy({top: dy, behavior: 'instant'});
const page = () => window.innerHeight - 2 * LINE;
const half = () => window.innerHeight / 2;

export function initKeys({openFile, reload}) {
    let lastG = 0;

    window.addEventListener('keydown', (e) => {
        if (search.isOpen()) return; // the search input handles its own keys
        if (e.isComposing) return;

        // While help is up, only the keys that close it do anything.
        if (help.isOpen()) {
            if (e.key === '?' || e.key === 'Escape' || e.key === 'q') help.close();
            e.preventDefault();
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

        if (e.key !== 'g') lastG = 0;
        switch (e.key) {
            case 'j': scrollBy(LINE); break;
            case 'k': scrollBy(-LINE); break;
            case 'd': scrollBy(half()); break;
            case 'u': scrollBy(-half()); break;
            case ' ':
            case 'f': scrollBy(e.shiftKey ? -page() : page()); break;
            case 'b': scrollBy(-page()); break;
            case 'G': window.scrollTo({top: document.body.scrollHeight, behavior: 'instant'}); break;
            case 'g': {
                const now = Date.now();
                if (now - lastG < GG_TIMEOUT) {
                    window.scrollTo({top: 0, behavior: 'instant'});
                    lastG = 0;
                } else {
                    lastG = now;
                }
                break;
            }
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
