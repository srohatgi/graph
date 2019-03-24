package graph_test

import (
	"context"
	"fmt"

	"github.com/srohatgi/graph"
)

type State int

const (
	Unknown State = iota
	Suspended
	Deleted
)

func (state State) String() string {
	switch state {
	case Suspended:
		return "Suspended"
	case Deleted:
		return "Deleted"
	}
	return "Unknown"
}

func FromString(from string) State {
	switch from {
	case "Suspended":
		return Suspended
	case "Deleted":
		return Deleted
	}
	return Unknown
}

const (
	numTenants int = 1
)

type Event struct {
	tenantID int
	state    State
}

type Stream []Event

func ProduceStream(tenantID int) Stream {
	var stream Stream
	for t := Suspended; t <= Deleted; t++ {
		stream = append(stream, Event{tenantID, t})
	}

	return stream
}

type Dispatcher struct {
	graph.Depends
	registrations []Service
}

func (stream Stream) Process(dispatcher *Dispatcher) {
	for _, evt := range stream {
		lib := graph.New(nil)

		resources := []graph.Resource{dispatcher}

		for _, r := range resources {
			r.SetEvent(evt.state.String())
		}

		toDelete := false

		if dispatcher.Event == "Deleted" {
			toDelete = true
		}

		lib.Sync(context.Background(), resources, toDelete)
	}
}

func (dispatcher *Dispatcher) Update(ctxt context.Context) (string, error) {
	for _, svc := range dispatcher.registrations {
		err := svc.Do(1, FromString(dispatcher.Depends.Event))
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

func (dispatcher *Dispatcher) Delete(ctxt context.Context) error {
	for _, svc := range dispatcher.registrations {
		err := svc.Do(1, FromString(dispatcher.Depends.Event))
		if err != nil {
			return err
		}
	}
	return nil
}

func (dispatcher *Dispatcher) Register(service Service) {
	dispatcher.registrations = append(dispatcher.registrations, service)
}

type Service interface {
	Do(tenantID int, state State) error
}

type service struct {
	name string
}

func (svc *service) Do(tenantID int, state State) error {
	fmt.Printf("processed %s for tenantID %d\n", state, tenantID)
	return nil
}

func Example_eventProcessing() {
	dispatcher := &Dispatcher{}
	dispatcher.Register(&service{name: "actions"})
	stream := ProduceStream(1)
	stream.Process(dispatcher)
	// Output:
	// processed Suspended for tenantID 1
	// processed Deleted for tenantID 1
}
