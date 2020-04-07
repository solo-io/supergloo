package description

import (
	"regexp"

	"github.com/rotisserie/eris"
)

const (
	AllPolicies     = "all"
	TrafficPolicies = "traffic"
	AccessPolicies  = "access"
)

var (
	InvalidResourceName = func(resourceName string) error {
		return eris.Errorf("Resource name %s is invalid - must be of the form \"name.namespace.registered-cluster-name\" (%s)", resourceName, resourceNamePattern)
	}

	// a resource name is three nonempty alphanumeric (plus - and _) fields separated by two dots
	resourceNamePattern = regexp.MustCompile("([a-zA-Z0-9_-]+)\\.([a-zA-Z0-9_-]+)\\.([a-zA-Z0-9_-]+)")
)

func ParseResourceName(resourceName string) (*FullyQualifiedKubeResource, error) {
	if !resourceNamePattern.MatchString(resourceName) {
		return nil, InvalidResourceName(resourceName)
	}

	matches := resourceNamePattern.FindStringSubmatch(resourceName)
	return &FullyQualifiedKubeResource{
		Name:        matches[1],
		Namespace:   matches[2],
		ClusterName: matches[3],
	}, nil
}
