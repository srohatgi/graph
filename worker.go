package graph

import (
	"sync"
)

type runnable interface {
	run()
}

// noOfWorkers for creating resources
var noOfWorkers = 3

// maxWorkQSize for outstanding runnables
var maxWorkQSize = 1000

var workQ = make(chan runnable, maxWorkQSize)
var wg = new(sync.WaitGroup)

func place(r runnable) {
	workQ <- r
}

func start() {
	// Adding routines to workgroup and running then
	for i := 0; i < noOfWorkers; i++ {
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
