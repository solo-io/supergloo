package resourceidutils

import (
	"fmt"
	"strings"

	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// Return a list of resourceIds as a string in the format "[ resource1Name.resource1Namespace.resource1Cluster, resource2Name.resource2Namespace.resource2Cluster,...]
func ResourceIdsToString(resourceIds []ezkube.ResourceId) string {
	var keys []string
	for _, resourceId := range resourceIds {
		keys = append(keys, sets.Key(resourceId))
	}
	return fmt.Sprintf("[%s]", strings.Join(keys, ", "))
}
