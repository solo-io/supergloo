package sets

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

func MatchMeshWorkloadSet(expected MeshWorkloadSet) types.GomegaMatcher {
	return &meshWorkloadSetMatcher{
		expected: expected,
	}
}

type meshWorkloadSetMatcher struct {
	expected MeshWorkloadSet
}

func (matcher *meshWorkloadSetMatcher) Match(actual interface{}) (success bool, err error) {
	meshWorkloadSet, ok := actual.(MeshWorkloadSet)
	if !ok {
		return false, fmt.Errorf("MatchMeshWorkloadSet expects a MeshWorkloadSet")
	}
	return matcher.expected.Equal(meshWorkloadSet), nil
}

func (matcher *meshWorkloadSetMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto equal the set \n\t%#v", actual, matcher.expected)
}

func (matcher *meshWorkloadSetMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to equal the set\n\t%#v", actual, matcher.expected)
}
