package sets

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

func MatchMeshServiceSet(expected MeshServiceSet) types.GomegaMatcher {
	return &meshServiceSetMatcher{
		expected: expected,
	}
}

type meshServiceSetMatcher struct {
	expected MeshServiceSet
}

func (matcher *meshServiceSetMatcher) Match(actual interface{}) (success bool, err error) {
	meshServiceSet, ok := actual.(MeshServiceSet)
	if !ok {
		return false, fmt.Errorf("MatchMeshServiceSet expects a MeshServiceSet")
	}
	return matcher.expected.Equal(meshServiceSet), nil
}

func (matcher *meshServiceSetMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto equal the set \n\t%#v", actual, matcher.expected)
}

func (matcher *meshServiceSetMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to equal the set\n\t%#v", actual, matcher.expected)
}
