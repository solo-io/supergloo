package matchers

import (
	"fmt"

	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
)

// Use this in a gomock EXPECT call e.g.
// `client.EXPECT().Update(ctx, GomockMatchPublicFields(expected)).Return(nil)`
func GomockMatchPublicFields(actual interface{}) gomock.Matcher {
	return &gomockPublicFieldMatcher{
		actual: actual,
	}
}

type gomockPublicFieldMatcher struct {
	actual interface{}
	diff   []string
}

func (p *gomockPublicFieldMatcher) Matches(actual interface{}) bool {
	diff := deep.Equal(p.actual, actual)
	p.diff = diff
	return len(diff) == 0
}

func (p *gomockPublicFieldMatcher) String() string {
	return fmt.Sprintf("equals proto %v", p.actual)
}
