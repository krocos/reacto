package reacto

import (
	"sync"

	"github.com/jtolio/gls"
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
	defer e.mu.Unlock()
	e.subscribers[handle] = struct{}{}
}

func (e *effects) notify() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for handle := range e.subscribers {
		wg.Add(1)
		h := handle // захват переменной
		gls.Go(func() {
			defer wg.Done()
			h.fn()
		})
	}
}

// ========== Управление эффектами ==========

var glsManager = gls.NewContextManager()

const effectName = "reacto-active-effect"

func withEffect(fn effect, run func()) {
	handle := &EffectHandle{fn: fn}
	glsManager.SetValues(gls.Values{
		effectName: handle,
	}, run)
}

func getActiveEffect() *EffectHandle {
	val, ok := glsManager.GetValue(effectName)
	if !ok {
		return nil
	}
	return val.(*EffectHandle)
}

// ========== Watch API ==========

func Watch(fn func()) {
	withEffect(fn, fn)
}

// ========== Синхронизация ==========

var wg sync.WaitGroup

// Wait блокирует до завершения всех активных эффектов
func Wait() {
	wg.Wait()
}
