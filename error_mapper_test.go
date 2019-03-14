package graph

import (
	"errors"
	"testing"
)

func TestErrorMap_Collection(t *testing.T) {
	f := func() error {
		em := errorMap{}
		em["0"] = errors.New("hello err 0")
		em["1"] = errors.New("hello err 1")
		return em
	}

	err := f()

	if err == nil {
		t.Fatalf("expected err to be not nil")
	}

	em, ok := err.(ErrorMapper)

	if !ok {
		t.Fatal("unable to get back ErrorMap")
	}

	errs := em.ErrorMap()

	for i, name := range []string{"0", "1"} {
		if _, ok := errs[name]; !ok {
			t.Fatalf("expected error to be present for index %d", i)
		}
	}
}
