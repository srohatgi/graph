package graph

import (
	"errors"
	"testing"
	"time"
)

type sleeper struct {
	t      *testing.T
	id     int
	tellme chan error
}

func (s *sleeper) run() {
	if s.id == 3 {
		s.t.Logf("Worker %d throwing error", s.id)
		s.tellme <- errors.New("issue in " + string(s.id))
	}
	time.Sleep(1 * time.Second)
	s.t.Logf("Worker %v done successfully", s.id)
}

func TestWorker(t *testing.T) {
	start()

	tasks := 7
	queue := make(chan error, tasks)

	for i := 0; i < tasks; i++ {
		s := &sleeper{t, i, queue}
		place(s)
	}

	defer stop()
}
