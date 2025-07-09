package reacto

import (
	"sync"

	"github.com/jtolio/gls"
)

type effect func()

type EffectHandle struct {
	fn effect
	wg *sync.WaitGroup
}

type WatchHandle struct {
	wg *sync.WaitGroup
}

func (w *WatchHandle) Wait() {
	if w != nil && w.wg != nil {
		w.wg.Wait()
	}
}

type ValueRef[T any] struct {
	mu      sync.RWMutex
	value   T
	effects *effects
}

// Ref creates new reference for a value.
func Ref[T any](value T) *ValueRef[T] {
	return &ValueRef[T]{
		value:   value,
		effects: newEffects(),
	}
}

// Value returns current value of ValueRef.
func (r *ValueRef[T]) Value() T {
	if h := getActiveEffect(); h != nil {
		r.effects.add(h)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Set sets new value for ValueRef and call all effects.
func (r *ValueRef[T]) Set(value T) {
	r.mu.Lock()
	r.value = value
	r.mu.Unlock()
	r.effects.notify()
}

type ComputedRef[T any] struct {
	compute func() T
}

// Computed creates new ComputedRef that computes everytime on Value method
// call.
func Computed[T any](compute func() T) *ComputedRef[T] {
	return &ComputedRef[T]{compute: compute}
}

func (c *ComputedRef[T]) Value() T {
	return c.compute()
}

type effects struct {
	mu          sync.Mutex
	subscribers map[*EffectHandle]struct{}
}

func newEffects() *effects {
	return &effects{
		subscribers: make(map[*EffectHandle]struct{}),
	}
}

func (e *effects) add(handle *EffectHandle) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subscribers[handle] = struct{}{}
}

func (e *effects) notify() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for handle := range e.subscribers {
		if handle.wg != nil {
			handle.wg.Add(1)
		}
		h := handle
		gls.Go(func() {
			defer func() {
				if h.wg != nil {
					h.wg.Done()
				}
			}()
			h.fn()
		})
	}
}

var glsManager = gls.NewContextManager()

const effectKey = "reacto-active-effect"

func withEffectHandle(fn effect, wg *sync.WaitGroup, run func()) {
	handle := &EffectHandle{
		fn: fn,
		wg: wg,
	}
	glsManager.SetValues(gls.Values{
		effectKey: handle,
	}, run)
}

func getActiveEffect() *EffectHandle {
	val, ok := glsManager.GetValue(effectKey)
	if !ok {
		return nil
	}
	return val.(*EffectHandle)
}

// Watch watches for all Refs in the callback and call it when any of Refs
// changed.
func Watch(fn func()) *WatchHandle {
	wg := &sync.WaitGroup{}
	withEffectHandle(fn, wg, fn)
	return &WatchHandle{wg: wg}
}

// WaitAll waits all handles to complete.
func WaitAll(handles ...*WatchHandle) {
	var wg sync.WaitGroup

	for _, h := range handles {
		if h != nil && h.wg != nil {
			wg.Add(1)
			go func(hw *sync.WaitGroup) {
				defer wg.Done()
				hw.Wait()
			}(h.wg)
		}
	}

	wg.Wait()
}

type WatchGroup struct {
	wg sync.WaitGroup
}

// NewWatchGroup creates new group of effects to wait to complete.
func NewWatchGroup() *WatchGroup {
	return &WatchGroup{}
}

// Add add another one WatchHandle to the group.
func (g *WatchGroup) Add(h *WatchHandle) {
	if h != nil && h.wg != nil {
		g.wg.Add(1)
		go func(hw *sync.WaitGroup) {
			defer g.wg.Done()
			hw.Wait()
		}(h.wg)
	}
}

// Wait waits for all WatchHandle to complete.
func (g *WatchGroup) Wait() {
	g.wg.Wait()
}
