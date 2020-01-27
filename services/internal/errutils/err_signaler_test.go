package errutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
)

var _ = Describe("err signaler", func() {
	var (
		errSignaler = &errorSignaler{
			errSignal: make(chan struct{}),
		}
	)
	It("will set an error", func() {
		testErr := eris.New("hello")
		errSignaler.SignalError(testErr)
		Expect(errSignaler.Error()).To(Equal(testErr))
	})
	It("will signal a channel when done", func() {
		testErr := eris.New("hello")
		errSignaler.SignalError(testErr)
		_, ok := <-errSignaler.GotError()
		Expect(ok).To(BeFalse())
		// make sure that the signal continue to be false, but catch properly
		_, ok = <-errSignaler.GotError()
		Expect(ok).To(BeFalse())
	})
})
