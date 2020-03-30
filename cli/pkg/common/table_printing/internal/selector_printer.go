package internal

import (
	"fmt"
	"strings"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
)

func WorkloadSelectorToCell(selector *core_types.WorkloadSelector) string {
	var namespaceField string
	if len(selector.GetNamespaces()) > 0 {
		namespaceField = fmt.Sprintf("Namespaces:\n%s\n", strings.Join(selector.GetNamespaces(), ","))
	}

	var labelsField string
	if len(selector.GetLabels()) > 0 {
		labelStrings := []string{}
		for k, v := range selector.GetLabels() {
			labelStrings = append(labelStrings, fmt.Sprintf("%s=%s", k, v))
		}
		labelsField = fmt.Sprintf("Labels:\n%s\n", strings.Join(labelStrings, ","))
	}

	return fmt.Sprintf("%s%s", namespaceField, labelsField)
}

func ServiceSelectorToCell(selector *core_types.ServiceSelector) string {
	switch selector.GetServiceSelectorType().(type) {
	case *core_types.ServiceSelector_Matcher_:
		namespaces := selector.GetMatcher().GetNamespaces()
		labels := selector.GetMatcher().GetLabels()
		clusters := selector.GetMatcher().GetClusters()

		var namespaceField string
		if len(namespaces) > 0 {
			namespaceField = fmt.Sprintf("Namespaces:\n%s\n", strings.Join(namespaces, ","))
		}

		var labelsField string
		if len(labels) > 0 {
			labelStrings := []string{}
			for k, v := range labels {
				labelStrings = append(labelStrings, fmt.Sprintf("%s=%s", k, v))
			}
			labelsField = fmt.Sprintf("Labels:\n%s\n", strings.Join(labelStrings, ","))
		}

		var clusterField string
		if len(clusters) > 0 {
			clusterField = fmt.Sprintf("Clusters:\n%s\n", strings.Join(clusters, "\n"))
		}

		return fmt.Sprintf("%s%s%s", namespaceField, labelsField, clusterField)
	case *core_types.ServiceSelector_ServiceRefs_:
		serviceRefs := selector.GetServiceRefs().GetServices()
		if len(serviceRefs) > 0 {
			var refs []string
			for _, ref := range serviceRefs {
				refs = append(refs, fmt.Sprintf("Name: %s\nNamespace: %s\nCluster: %s", ref.GetName(), ref.GetNamespace(), ref.GetCluster()))
			}

			return strings.Join(refs, "\n")
		}
	default:
		return fmt.Sprintf("Unknown selector type: %+v", selector)
	}
	return ""
}
