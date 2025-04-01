package reacto

import (
	"sync"
	"unsafe"
)

type effect func()

var activeEffect effect

// ---------------------------------------------------------------------------- //

type ValueRef[T any] struct {
	effects *effects
	value   T
}

func Ref[T any](value T) *ValueRef[T] {
	return &ValueRef[T]{
		effects: newEffects(),
		value:   value,
	}
}

func (r *ValueRef[T]) Value() T {
	r.effects.add(activeEffect)
	return r.value
}

func (r *ValueRef[T]) Set(value T) {
	r.value = value
	r.effects.notify()
}

// ---------------------------------------------------------------------------- //

var mu sync.Mutex

func Watch(e effect) {
	mu.Lock()
	defer mu.Unlock()

	activeEffect = e
	activeEffect()
	activeEffect = nil
}

// ---------------------------------------------------------------------------- //
// Вычисляемая переменная. При создании принимает функцию, которая вычисляет значение.

type ComputedRef[T any] struct {
	//effects *effects
	compute func() T
}

func Computed[T any](compute func() T) *ComputedRef[T] {
	return &ComputedRef[T]{
		//effects: newEffects(),
		compute: compute,
	}
}

func (c *ComputedRef[T]) Value() T {
	//c.effects.add(activeEffect)
	return c.compute()
}

// ---------------------------------------------------------------------------- //
// Набор эффектов для выполнения и добавления.

type effects struct {
	subscribers sync.Map
}

func newEffects() *effects {
	return &effects{subscribers: sync.Map{}}
}

func (e *effects) add(ef effect) {
	if ef == nil {
		return
	}

	key := uintptr(unsafe.Pointer(&ef))
	e.subscribers.Store(key, activeEffect)
}

func (e *effects) notify() {
	e.subscribers.Range(func(_, ef any) bool {
		ef.(effect)()
		return true
	})
}
