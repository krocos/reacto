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

	w := reacto.Watch(func() {
		t.Log("revenue:", revenue.Value())
	})

	price.Set(3)
	quantity.Set(3)

	if revenue.Value() != 9 {
		t.Fatal("unexpected result")
	}

	w.Wait()
}

func TestCorrectReactions(t *testing.T) {
	a := reacto.Ref(1)
	b := reacto.Ref(1)
	c := reacto.Ref(1)

	sumAB := reacto.Computed(func() int { // 2
		return a.Value() + b.Value()
	})

	sumAC := reacto.Computed(func() int { // 2
		return a.Value() + c.Value()
	})

	w1 := reacto.Watch(func() {
		v := sumAB.Value()
		t.Log("a + b:", v)
	})

	w2 := reacto.Watch(func() {
		v := sumAC.Value()
		t.Log("a + c:", v)
	})

	a.Set(2) // ab == 3 ac == 3
	if sumAB.Value() != 3 {
		t.Error("unexpected result")
	}
	if sumAC.Value() != 3 {
		t.Error("unexpected result")
	}
	a.Set(3) // ab == 4 ac == 4
	if sumAB.Value() != 4 {
		t.Error("unexpected result")
	}
	if sumAC.Value() != 4 {
		t.Error("unexpected result")
	}

	reacto.WaitAll(w1, w2)
}

type User struct {
	Name  string
	Age   int
	Phone string
}

type Card struct {
	Number *reacto.ValueRef[string]
	Pin    string
}

type State struct {
	User      *reacto.ValueRef[*User]
	Card      *reacto.ValueRef[*Card]
	CardTitle *reacto.ComputedRef[string]
}

func TestReal(t *testing.T) {
	s := State{
		User: reacto.Ref(&User{
			Name:  "User",
			Age:   10,
			Phone: "123",
		}),
		Card: reacto.Ref(&Card{
			Number: reacto.Ref("123123"),
			Pin:    "456",
		}),
	}

	s.CardTitle = reacto.Computed(func() string {
		return s.User.Value().Name + s.Card.Value().Number.Value()
	})

	w := reacto.NewWatchGroup()

	w.Add(reacto.Watch(func() {
		t.Log(s.Card.Value())
	}))

	w.Add(reacto.Watch(func() {
		t.Log(s.User.Value())
	}))

	card := s.Card.Value()

	w.Add(reacto.Watch(func() {
		t.Log(card.Number.Value())
	}))

	w.Add(reacto.Watch(func() {
		t.Log(s.CardTitle.Value())
	}))

	t.Log("-------------")

	card.Pin = "090900909"
	s.Card.Set(card)

	t.Log("-------------")

	card.Number.Set("0999009990")

	w.Wait()
}

func TestWatchFor(t *testing.T) {
	v1 := reacto.Ref(1)
	v2 := reacto.Ref(2)
	v3 := reacto.Ref(3)

	c1 := reacto.Computed(func() int {
		return v1.Value() + v2.Value()
	})

	c2 := reacto.Computed(func() int {
		return v1.Value() + v3.Value()
	})

	w1 := reacto.Watch(func() {
		t.Log("v1+v2:", v1.Value()+v2.Value())
	})

	w2 := reacto.Watch(func() {
		t.Log("v1+v3:", v1.Value()+v3.Value())
	})

	if c1.Value() != 3 {
		t.Error("unexpected c1 value")
	}

	if c2.Value() != 4 {
		t.Error("unexpected c2 value")
	}

	v1.Set(2)

	if c1.Value() != 4 {
		t.Error("unexpected c1 value")
	}

	if c2.Value() != 5 {
		t.Error("unexpected c2 value")
	}

	reacto.WaitAll(w1, w2)
}
