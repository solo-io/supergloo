package internal

import (
	"fmt"
	"strings"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
)

func SelectorToCell(selector *core_types.Selector) string {
	if len(selector.Refs) > 0 {
		var refs []string
		for _, ref := range selector.Refs {
			refs = append(refs, fmt.Sprintf("Name: %s\nNamespace: %s\nCluster: %s", ref.GetName(), ref.GetNamespace(), ref.GetCluster().GetValue()))
		}

		return strings.Join(refs, "\n")
	}

	var namespaceField string
	if len(selector.Namespaces) > 0 {
		namespaceField = fmt.Sprintf("Namespaces:\n%s\n", strings.Join(selector.Namespaces, ","))
	}

	var labelsField string
	if len(selector.Labels) > 0 {
		labels := []string{}
		for k, v := range selector.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}
		labelsField = fmt.Sprintf("Labels:\n%s\n", strings.Join(labels, ","))
	}

	var clusterField string
	if selector.GetCluster().GetValue() != "" {
		clusterField = fmt.Sprintf("Cluster: %s\n", selector.Cluster.Value)
	}

	return fmt.Sprintf("%s%s%s", namespaceField, labelsField, clusterField)
}
