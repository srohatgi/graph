package graph

import (
	"errors"
	"testing"
)

func TestErrorMap_NilIsNil(t *testing.T) {
	var es ErrorMap
	var err error

	err = es.Get()

	if err != nil {
		t.Fatalf("expected es: %v, to be nil", err)
	}
}

func TestErrorMap_Collection(t *testing.T) {
	f := func() error {
		em := ErrorMap{}
		em["0"] = errors.New("hello err 0")
		em["1"] = errors.New("hello err 1")
		return em.Get()
	}

	err := f()

	if err == nil {
		t.Fatalf("expected err to be not nil")
	}

	emPtr, ok := err.(*ErrorMap)

	if !ok {
		t.Fatal("unable to get back ErrorMap")
	}

	if emPtr.Size() != 2 {
		t.Fatalf("got size as %d, expected 2", emPtr.Size())
	}

	em := *emPtr
	for i, name := range []string{"0", "1"} {
		if _, ok := em[name]; !ok {
			t.Fatalf("expected error to be present for index %d", i)
		}
	}
}
