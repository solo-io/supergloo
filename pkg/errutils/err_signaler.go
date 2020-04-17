package errutils

import (
	"sync"
)

type ErrorSignaler interface {
	SignalError(err error)
	Error() error
	GotError() <-chan struct{}
}

func NewErrorSignaler() *errorSignaler {
	return &errorSignaler{
		errSignal: make(chan struct{}),
	}
}

type errorSignaler struct {
	// errSignal indicates that an error occurred, when closed.  It shouldn't
	// be written to.
	errSignal chan struct{}

	// err is the received error
	err error

	mu sync.Mutex
}

func (r *errorSignaler) SignalError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err == nil {
		// non-error, ignore
		return
	}

	if r.err != nil {
		// we already have an error, don't try again
		return
	}

	// save the error and report it
	r.err = err
	close(r.errSignal)
}

func (r *errorSignaler) Error() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.err
}

func (r *errorSignaler) GotError() <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.errSignal
}
