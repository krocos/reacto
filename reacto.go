package reacto

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
)

// ========== Типы ==========

type effect func()

type EffectHandle struct {
	fn effect
}

// ========== Ref ==========

type ValueRef[T any] struct {
	mu      sync.RWMutex
	value   T
	effects *effects
}

func Ref[T any](value T) *ValueRef[T] {
	return &ValueRef[T]{
		value:   value,
		effects: newEffects(),
	}
}

func (r *ValueRef[T]) Value() T {
	if h := getActiveEffect(); h != nil {
		r.effects.add(h)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

func (r *ValueRef[T]) Set(value T) {
	r.mu.Lock()
	r.value = value
	r.mu.Unlock()
	r.effects.notify()
}

// ========== Computed ==========

type ComputedRef[T any] struct {
	compute func() T
}

func Computed[T any](compute func() T) *ComputedRef[T] {
	return &ComputedRef[T]{compute: compute}
}

func (c *ComputedRef[T]) Value() T {
	return c.compute()
}

// ========== Эффекты ==========

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
	e.subscribers[handle] = struct{}{}
	e.mu.Unlock()
}

func (e *effects) notify() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for handle := range e.subscribers {
		go handle.fn()
	}
}

// ========== Управление эффектами ==========

var activeEffectStorage sync.Map // map[goroutine-id]*EffectHandle

func withEffect(fn effect, run func()) {
	handle := &EffectHandle{fn: fn}
	setActiveEffect(handle)
	run()
	clearActiveEffect()
}

func setActiveEffect(h *EffectHandle) {
	activeEffectStorage.Store(getGID(), h)
}

func getActiveEffect() *EffectHandle {
	if val, ok := activeEffectStorage.Load(getGID()); ok {
		return val.(*EffectHandle)
	}
	return nil
}

func clearActiveEffect() {
	activeEffectStorage.Delete(getGID())
}

// ========== Watch API ==========

func Watch(fn effect) {
	withEffect(fn, fn)
}

// ========== Получение goroutine ID (неофициально!) ==========

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	i := bytes.IndexByte(b, ' ')
	id, _ := strconv.ParseUint(string(b[:i]), 10, 64)
	return id
}
