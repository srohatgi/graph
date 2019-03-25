package graph_test

import (
	"context"
	"fmt"

	"github.com/srohatgi/graph"
)

// A small example to illustrate event processing on a tenant CRD.
// The tenant CRD may contain a reference to a Dispatcher Resource (see below).
//
// An external event may be communicated to a CRD by creating a label (Dispatcher, Suspended);
// which can then be used by the Dispatcher to propogate to its registered Service collection.
//
// The registration of a Service to a Dispatcher would be out of band; one
// example may be to use a Kubernetes ConfigMap loaded as a volume.
func Example_eventProcessing() {
	dispatcher := &Dispatcher{}
	dispatcher.Register(&service{name: "actions"})

	stream := ProduceStream(1)
	stream.Process([]graph.Resource{dispatcher})
	// Output:
	// processed Suspended for tenantID 1
	// processed Resumed for tenantID 1
	// processed Deleted for tenantID 1
}

// State captures events on a tenant, an entity in the system.
type State int

// Valid event states.
const (
	Unknown State = iota
	Suspended
	Resumed
	Deleted
)

func (state State) String() string {
	switch state {
	case Resumed:
		return "Resumed"
	case Suspended:
		return "Suspended"
	case Deleted:
		return "Deleted"
	}
	return "Unknown"
}

// FromString is a convenience function.
func FromString(from string) State {
	switch from {
	case "Resumed":
		return Resumed
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

// Event codifies tenant to state relationship.
type Event struct {
	tenantID int
	state    State
}

// Stream of events.
type Stream []Event

// ProduceStream simulates events for a given tenant.
func ProduceStream(tenantID int) Stream {
	var stream Stream
	for t := Suspended; t <= Deleted; t++ {
		stream = append(stream, Event{tenantID, t})
	}

	return stream
}

// Process events on a set of resources.
func (stream Stream) Process(resources []graph.Resource) {
	for _, evt := range stream {
		lib := graph.New(nil)

		for _, r := range resources {
			r.SetEvent(evt.state.String())
		}

		lib.Sync(context.Background(), resources, evt.state == Deleted)
	}
}

// Dispatcher propogates events to services.
type Dispatcher struct {
	graph.Depends
	registrations []Service
}

// Update fulfills graph.Resource interface.
func (dispatcher *Dispatcher) Update(ctxt context.Context) (string, error) {
	for _, svc := range dispatcher.registrations {
		err := svc.Do(1, FromString(dispatcher.Depends.Event))
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

// Delete fulfills graph.Resource interface.
func (dispatcher *Dispatcher) Delete(ctxt context.Context) error {
	for _, svc := range dispatcher.registrations {
		err := svc.Do(1, FromString(dispatcher.Depends.Event))
		if err != nil {
			return err
		}
	}
	return nil
}

// Register any arbitrary service.
func (dispatcher *Dispatcher) Register(service Service) {
	dispatcher.registrations = append(dispatcher.registrations, service)
}

// Service is an abstract interface for a real service.
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
