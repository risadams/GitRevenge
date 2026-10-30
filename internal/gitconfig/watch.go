package gitconfig

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/syncthing/notify"
)

// watchDebounce is how long a file must be quiet before a change is reported.
const watchDebounce = 150 * time.Millisecond

// Watcher reports changes to individual config files. Repo watchers don't cover global or
// included config, which can live anywhere. It watches each file's folder (not
// recursively), since editors and git replace config files by renaming over them.
type Watcher struct {
	onChange func(path string)
	events   chan notify.EventInfo
	done     chan struct{}
	wg       sync.WaitGroup

	mu sync.Mutex
	// files maps a normalized path to the path reported for it.
	files map[string]string
	dirs  map[string]bool
}

// NewWatcher starts a Watcher that calls onChange, from its own goroutine, with the path
// given to Add.
func NewWatcher(onChange func(path string)) *Watcher {
	w := &Watcher{
		onChange: onChange,
		events:   make(chan notify.EventInfo, 256),
		done:     make(chan struct{}),
		files:    map[string]string{},
		dirs:     map[string]bool{},
	}
	w.wg.Add(1)
	go w.loop()
	return w
}

// Add watches path. A symlinked file is watched at both ends. Folders that don't exist are
// skipped; the file can't change until something creates them.
func (w *Watcher) Add(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	targets := []string{filepath.Clean(path)}
	if r, err := filepath.EvalSymlinks(path); err == nil && r != targets[0] {
		targets = append(targets, r)
	}
	for _, t := range targets {
		w.files[Normalize(t)] = path
		dir := filepath.Dir(t)
		if w.dirs[dir] {
			continue
		}
		if err := notify.Watch(dir, w.events, notify.All); err == nil {
			w.dirs[dir] = true
		}
	}
}

// Close stops watching.
func (w *Watcher) Close() {
	notify.Stop(w.events)
	close(w.done)
	w.wg.Wait()
}

func (w *Watcher) loop() {
	defer w.wg.Done()
	pending := map[string]bool{}
	var timer <-chan time.Time
	for {
		select {
		case <-w.done:
			return
		case ev := <-w.events:
			w.mu.Lock()
			p, ok := w.files[Normalize(ev.Path())]
			w.mu.Unlock()
			if ok {
				pending[p] = true
				timer = time.After(watchDebounce)
			}
		case <-timer:
			for p := range pending {
				w.onChange(p)
			}
			pending = map[string]bool{}
			timer = nil
		}
	}
}
