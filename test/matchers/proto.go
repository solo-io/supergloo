package matchers

import (
	"fmt"
	"strings"

	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
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
	return fmt.Sprintf("%+v", p.actual)
}

func (p *gomockPublicFieldMatcher) Got(got interface{}) string {

	if interfaceList, ok := got.([]interface{}); ok {
		var items = []string{"Items:"}
		for _, v := range interfaceList {
			items = append(items, fmt.Sprintf("%+v", v))
		}
		return strings.Join(items, "\n")
	}
	return fmt.Sprintf("%+v", got)
}

func MatchPublicFields(obj interface{}) types.GomegaMatcher {
	return &publicFieldMatcherImpl{
		obj: obj,
	}
}

type publicFieldMatcherImpl struct {
	obj interface{}
}

func (p *publicFieldMatcherImpl) Match(actual interface{}) (success bool, err error) {
	diff := deep.Equal(p.obj, actual)
	return len(diff) == 0, nil
}

func (p *publicFieldMatcherImpl) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "To be identical to", p.obj)
}

func (p *publicFieldMatcherImpl) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "Not to be identical to", p.obj)
}
