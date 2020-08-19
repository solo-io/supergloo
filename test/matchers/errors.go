package matchers

import (
	"github.com/golang/mock/gomock"
)

func MatchesError(e error) gomock.Matcher {
	return &errorMatcher{err: e}
}

type errorMatcher struct {
	err error
}

func (e *errorMatcher) Matches(x interface{}) bool {
	err, ok := x.(error)
	if !ok {
		return false
	}
	return err.Error() == e.err.Error()
}

func (e *errorMatcher) String() string {
	return e.err.Error()
}

