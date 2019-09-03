// Package chantest implements utilities for testing concurrency.
package chantest

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// Default shold be plenty of time for a goroutine to reach a send to a channel
// if not blocked or doing something slow.
const Default = Before(100 * time.Millisecond)

// Expect calls Before.Expect on Default.
func Expect(t *testing.T, do func()) {
	Default.Expect(t, do)
}

// AssertRecv calls Before.AssertRecv on Default.
func AssertRecv(t *testing.T, ch interface{}, msgAndArgs ...interface{}) interface{} {
	return Default.AssertRecv(t, ch, msgAndArgs...)
}

// AssertNoRecv calls Before.AssertNoRecv on Default.
func AssertNoRecv(t *testing.T, ch interface{}, msgAndArgs ...interface{}) interface{} {
	return Default.AssertNoRecv(t, ch, msgAndArgs...)
}

// AssertSend calls Before.AssertSend on Default.
func AssertSend(t *testing.T, ch, v interface{}, msgAndArgs ...interface{}) {
	Default.AssertSend(t, ch, v, msgAndArgs...)
}

// AssertNoSend calls Before.AssertNoSend on Default.
func AssertNoSend(t *testing.T, ch, v interface{}, msgAndArgs ...interface{}) {
	Default.AssertNoSend(t, ch, v, msgAndArgs...)
}

// Before is the amount of time to wait before failing an expectation.
type Before time.Duration

// Expect fails the test if do doesn't return very quickly, typically after
// blocking for a single channel operation.
//
// Useful for testing that a goroutine is unblocked and has reached a certain
// point that somehow reads or sends to the do function, and to synchronize
// its continuation with
func (d Before) Expect(t *testing.T, do func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		do()
	}()
	select {
	case <-done:
	case <-time.After(time.Duration(d)):
		t.Fatal("timeout waiting for channel send or receive")
	}
}

// AssertRecv asserts that something is quickly received from ch, which must be a channel.
// custom msgAndArgs cand be added, with first argument being the formatted string
func (d Before) AssertRecv(t *testing.T, ch interface{}, msgAndArgs ...interface{}) interface{} {
	t.Helper()
	v, ok := d.assertRecv(t, ch)
	if !ok {
		t.Fatal(defaultOrCustomMessage("timeout waiting for channel send or receive", msgAndArgs...))
	}
	return v
}

// AssertNoRecv asserts that nothing is received from ch, which must be a channel, for a very short period of time.
// custom msgAndArgs cand be added, with first argument being the formatted string
func (d Before) AssertNoRecv(t *testing.T, ch interface{}, msgAndArgs ...interface{}) interface{} {
	t.Helper()
	v, ok := d.assertRecv(t, ch)
	if !ok {
		return nil
	}
	t.Fatal(defaultOrCustomMessage("unexpected channel receive", msgAndArgs...))
	return v
}

// AssertSend asserts that v is quickly sent from ch, which must be a channel.
// custom msgAndArgs cand be added, with first argument being the formatted string
func (d Before) AssertSend(t *testing.T, ch, v interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !d.assertSend(t, ch, v) {
		t.Fatal(defaultOrCustomMessage("timeout waiting for channel send or receive", msgAndArgs...))
	}
}

// AssertNoSend asserts that v is not sent to ch, which must be a channel, for a very short period of time.
// custom msgAndArgs cand be added, with first argument being the formatted string
func (d Before) AssertNoSend(t *testing.T, ch, v interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if d.assertSend(t, ch, v) {
		t.Fatal(defaultOrCustomMessage("unexpected channel receive", msgAndArgs...))
	}
}

func (d Before) assertRecv(t *testing.T, ch interface{}) (interface{}, bool) {
	t.Helper()

	// lol no generics
	//
	// var ch <-chan T
	// var v T
	// select {
	// case v = <-ch:
	//    chosen = 0
	// case <-time.After(time.Duration(d)):
	//    chosen = 1
	// }
	chosen, recv, _ := reflect.Select([]reflect.SelectCase{{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ch),
	}, {
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(time.After(time.Duration(d))),
	}})
	if chosen != 0 {
		return nil, false
	}

	return recv.Interface(), true
}

func (d Before) assertSend(t *testing.T, ch, v interface{}) bool {
	t.Helper()

	// lol no generics
	//
	// var ch <-chan T
	// var v T
	// select {
	// case ch <- v:
	//    chosen = 0
	// case <-time.After(time.Duration(d)):
	//    chosen = 1
	// }
	chosen, _, _ := reflect.Select([]reflect.SelectCase{{
		Chan: reflect.ValueOf(ch),
		Dir:  reflect.SelectSend,
		Send: reflect.ValueOf(v),
	}, {
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(time.After(time.Duration(d))),
	}})

	return chosen == 0
}

// defaultOrCustomMessage tries to format customMsgAndArgs as format string and optional args,
// if the formatting returns a non-empty string, it returns it, otherwise returns defaultMessage
func defaultOrCustomMessage(defaultMessage string, customMsgAndArgs ...interface{}) string {
	msg := messageFromMsgAndArgs(customMsgAndArgs...)
	if msg == "" {
		return defaultMessage
	}
	return msg
}

// copied from testify/assert/assertions.go
func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		return msgAndArgs[0].(string)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
