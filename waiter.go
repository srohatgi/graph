package graph

// This waiter is inspired by AWS request waiter.

import (
"context"
"fmt"
"time"
)


// A Waiter provides the functionality to perform a blocking call which will
// wait for a resource state to be satisfied by a service.
// Internally, we will stop retry the requests if we can satisfy one of the Acceptors
type Waiter struct {
    Acceptors []WaiterAcceptor

    MaxAttempts int
    Delay       time.Duration

    ExecuteAction    func() interface{}
    SleepWithContext func(context.Context, time.Duration) error
}

// A WaiterAcceptor provides the information needed to wait for an API operation
// to complete.
type WaiterAcceptor struct {
    Matcher func(interface{}) bool
}

func (a *WaiterAcceptor) match(ctx context.Context, i interface{}) bool {
    return a.Matcher(i)
}

// WaitWithContext will call the ExecuteAction() internally. The request's response will be compared against the
// Waiter's Acceptors to determine the successful state of the resource the
// waiter is inspecting.
//
// The passed in context must not be nil. If it is nil a panic will occur. The
// Context will be used to cancel the waiter's pending requests and retry delays.
//
// The waiter will continue until the target state defined by the Acceptors,
// or the max attempts expires.
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

