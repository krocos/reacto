package reacto

import (
	"strings"
	"sync"
)

type effect func()

var (
	activeToken  string
	activeEffect effect
)

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
	r.effects.add()
	return r.value
}

func (r *ValueRef[T]) Set(value T) {
	if hasActiveEffect() {
		panic("using Set() method on Ref[T] in Watch() function")
	}

	r.value = value
	r.effects.notify()
}

var mu sync.Mutex

func Watch(token string, e effect) {
	mu.Lock()
	defer func() {
		activeToken = ""
		activeEffect = nil
		mu.Unlock()
	}()

	activeToken = token
	activeEffect = e
	activeEffect()
}

type ComputedRef[T any] struct {
	compute func() T
}

func Computed[T any](compute func() T) *ComputedRef[T] {
	return &ComputedRef[T]{
		compute: compute,
	}
}

func (c *ComputedRef[T]) Value() T {
	return c.compute()
}

type effects struct {
	subscribers sync.Map
}

func newEffects() *effects {
	return &effects{subscribers: sync.Map{}}
}

func (e *effects) add() {
	if !hasActiveEffect() {
		return
	}

	e.subscribers.Store(activeToken, activeEffect)
}

func (e *effects) notify() {
	e.subscribers.Range(func(_, ef any) bool {
		ef.(effect)()
		return true
	})
}

func hasActiveEffect() bool {
	return strings.TrimSpace(activeToken) != "" && activeEffect != nil
}
