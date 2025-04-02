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

	reacto.Watch("token", func() {
		t.Log("revenue:", revenue.Value())
	})

	price.Set(3)
	quantity.Set(3)

	if revenue.Value() != 9 {
		t.Fatal("unexpected result")
	}
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

	reacto.Watch("here we watch for something", func() {
		v := sumAB.Value()
		t.Log("a + b:", v)
	})

	reacto.Watch("here we log something", func() {
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

	reacto.Watch("card data changed", func() {
		t.Log("Card changed", s.Card.Value())
	})

	reacto.Watch("user data changed", func() {
		t.Log("New user data", s.User.Value())
	})

	card := s.Card.Value()

	reacto.Watch("card number changed", func() {
		t.Log("Card num changed", card.Number.Value())
	})

	reacto.Watch("CARD TITLE CHANGED", func() {
		t.Log("CARD TITLE CHANGED", s.CardTitle.Value())
	})

	t.Log("-------------")

	card.Pin = "090900909"
	s.Card.Set(card)

	t.Log("-------------")

	card.Number.Set("0999009990")
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()

	a := reacto.Ref(1)
	reacto.Watch("a", func() {
		a.Set(3)
	})
}
