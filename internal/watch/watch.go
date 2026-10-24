// Package watch reports changes to a single file.
package watch

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounce = 150 * time.Millisecond

// Watcher calls a function shortly after its file changes.
type Watcher struct {
	fs   *fsnotify.Watcher
	done chan struct{}
	once sync.Once
}

// File watches path and calls onChange after writes settle. It watches the
// parent directory because many editors save by writing a temp file and
// renaming it over the original, which drops a watch on the file itself.
func File(path string, onChange func()) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fw.Add(filepath.Dir(path)); err != nil {
		fw.Close()
		return nil, err
	}
	w := &Watcher{fs: fw, done: make(chan struct{})}
	go w.loop(filepath.Clean(path), onChange)
	return w, nil
}

func (w *Watcher) loop(path string, onChange func()) {
	var timer *time.Timer
	for {
		select {
		case <-w.done:
			if timer != nil {
				timer.Stop()
			}
			return
		case ev, ok := <-w.fs.Events:
			if !ok {
				return
			}
			if filepath.Clean(ev.Name) != path || !ev.Has(fsnotify.Write|fsnotify.Create|fsnotify.Rename) {
				continue
			}
			if timer == nil {
				timer = time.AfterFunc(debounce, onChange)
			} else {
				timer.Reset(debounce)
			}
		case _, ok := <-w.fs.Errors:
			if !ok {
				return
			}
		}
	}
}

// Close stops watching. It is safe to call more than once.
func (w *Watcher) Close() error {
	var err error
	w.once.Do(func() {
		close(w.done)
		err = w.fs.Close()
	})
	return err
}
