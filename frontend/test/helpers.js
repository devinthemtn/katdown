// Sends a keydown to target (default: the window) and returns the event.
export function key(k, opts = {}, target = window) {
    const e = new KeyboardEvent('keydown', {key: k, bubbles: true, cancelable: true, ...opts});
    target.dispatchEvent(e);
    return e;
}

// Waits for pending promises (and the microtasks they queue) to settle.
export const flush = () => new Promise((r) => setTimeout(r, 0));
