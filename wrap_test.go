package graph

import "testing"

func TestErrorSlice(t *testing.T) {
	var es ErrorSlice
	var err error

	err = es.Get()

	if err != nil {
		t.Fatalf("expected es: %v, to be nil", err)
	}
}
