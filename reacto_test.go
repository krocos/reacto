package reacto_test

import (
	"testing"

	"github.com/krocos/reacto"
)

func TestReacto(t *testing.T) {
	price := reacto.Ref(2)
	quantity := reacto.Ref(2)

	revenue := reacto.Computed(func() int {
		return price.Value() * quantity.Value()
	})

	reacto.Watch(func() {
		t.Log("revenue:", revenue.Value())
	})

	price.Set(3)
	quantity.Set(3)

	if revenue.Value() != 9 {
		t.Fatal("unexpected result")
	}
}
