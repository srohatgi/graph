package graph

import (
	"fmt"
	"testing"
	"time"
)

type sleeper struct {
	id     int
	tellme chan error
}

func (s *sleeper) run() {
	time.Sleep(1 * time.Second)
	fmt.Printf("Done with id: %v\n", s.id)
}

func TestWorker(t *testing.T) {
	start()

	tasks := 7
	queue := make(chan error, tasks)

	for i := 0; i < tasks; i++ {
		s := &sleeper{i, queue}
	}

	defer stop()
}
