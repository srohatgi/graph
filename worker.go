package graph

import (
	"sync"
)

type runnable interface {
	run()
}

// NoOfWorkers for creating resources
var NoOfWorkers = 3

// MaxWorkQSize for outstanding runnables
var MaxWorkQSize = 1000

var workQ = make(chan runnable, MaxWorkQSize)
var wg = new(sync.WaitGroup)

func place(r runnable) {
	workQ <- r
}

func start() {
	// Adding routines to workgroup and running then
	for i := 0; i < NoOfWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workQ {
				work.run()
			}
		}()
	}
}

func stop() {
	close(workQ)
	wg.Wait()
}
