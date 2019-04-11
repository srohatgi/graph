package graph

import (
	"context"
	"fmt"
	"time"
)

// Waiter provides a generic wrapper for polling readiness
// of a Resource. Design is inspired by AWS request waiter.
type Waiter struct {
	Acceptors []WaiterAcceptor

	MaxAttempts int
	Delay       time.Duration

	ExecuteAction    func() interface{}
	SleepWithContext func(context.Context, time.Duration) error
}

// WaiterAcceptor performs the readiness check.
type WaiterAcceptor struct {
	Matcher func(interface{}) bool
}

func (a *WaiterAcceptor) match(ctx context.Context, i interface{}) bool {
	return a.Matcher(i)
}

// WaitWithContext calls the ExecuteAction() internally. The request's response will be matched
// with the waiter's Acceptors to determine the "ready" state of the resource the waiter is inspecting.
//
// The passed in Context must not be nil. If it is nil a panic will occur. The
// Context will be used to cancel the waiter's pending requests and retry delays.
//
// The function continues till the Resource is ready OR MaxAttempts is reached.
func (w Waiter) WaitWithContext(ctx context.Context) error {

	for attempt := 1; ; attempt++ {
		//execute
		i := w.ExecuteAction()

		// See if any of the acceptors match the request's response, or error
		for _, a := range w.Acceptors {
			if matched := a.match(ctx, i); matched {
				return nil
			}
		}

		// The Waiter should only check the resource state MaxAttempts times
		// This is here instead of in the for loop above to prevent delaying
		// unnecessary when the waiter will not retry.
		if attempt == w.MaxAttempts {
			break
		}

		sleepCtxFn := w.SleepWithContext
		if sleepCtxFn == nil {
			sleepCtxFn = sleepWithContext
		}

		if err := sleepCtxFn(ctx, w.Delay); err != nil {
			return fmt.Errorf("waiter context canceled")
		}

	}

	return fmt.Errorf("exceeded wait attempts")
}

// sleepWithContext will wait for the timer duration to expire, or the context
// is canceled. Which ever happens first. If the context is canceled the Context's
// error will be returned.
//
// Expects Context to always return a non-nil error if the Done channel is closed.
func sleepWithContext(ctx context.Context, dur time.Duration) error {
	t := time.NewTimer(dur)
	defer t.Stop()

	select {
	case <-t.C:
		break
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
