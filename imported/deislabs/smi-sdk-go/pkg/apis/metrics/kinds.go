package metrics

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var AvailableKinds = map[string]*metav1.APIResource{
	"Deployment": {
		Name:       "deployments",
		Namespaced: true,
		Kind:       "TrafficMetrics",
		Verbs: []string{
			"get",
			"list",
		},
	},
	"Pod": {
		Name:       "pods",
		Namespaced: true,
		Kind:       "TrafficMetrics",
		Verbs: []string{
			"get",
			"list",
		},
	},
	"Daemonset": {
		Name:       "daemonsets",
		Namespaced: true,
		Kind:       "TrafficMetrics",
		Verbs: []string{
			"get",
			"list",
		},
	},
	"Statefulset": {
		Name:       "statefulsets",
		Namespaced: true,
		Kind:       "TrafficMetrics",
		Verbs: []string{
			"get",
			"list",
		},
	},
	"Namespace": {
		Name:       "namespaces",
		Namespaced: false,
		Kind:       "TrafficMetrics",
		Verbs: []string{
			"get",
			"list",
		},
	},
}

func getKindName(kind string) string {
	apiResource, ok := AvailableKinds[kind]
	if !ok {
		return "unsupported"
	}

	return apiResource.Name
}
