package metrics

import (
	"fmt"
	"path"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func uniqueSelfLink(obj *v1.ObjectReference) string {
	// If Namespace is empty, it is assumed that this is a non-namespaced resource.
	if obj.Namespace == "" {
		return path.Join(baseURL(), getKindName(obj.Kind), obj.Name)
	}

	return path.Join(
		baseURL(),
		"namespaces",
		obj.Namespace,
		getKindName(obj.Kind),
		obj.Name)
}

// NewTrafficMetrics constructs a TrafficMetrics with all the defaults
func NewTrafficMetrics(obj, edge *v1.ObjectReference) *TrafficMetrics {
	selfLink := uniqueSelfLink(obj)

	if edge != nil {
		selfLink = path.Join(selfLink, "edges")
	}

	metrics := []*Metric{}
	for _, m := range AvailableMetrics {
		n := *m
		metrics = append(metrics, &n)
	}

	resource := &TrafficMetrics{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrafficMetrics",
			APIVersion: APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
			Name:              obj.Name,
			Namespace:         obj.Namespace,
			SelfLink:          selfLink,
		},
		Resource: obj,
		Metrics:  metrics,
	}

	if edge != nil {
		resource.Edge = &Edge{
			Resource: edge,
		}
	}

	return resource
}

// TrafficMetrics provide the metrics for a specific resource
type TrafficMetrics struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	*Interval `json:",inline"`

	Resource *v1.ObjectReference `json:"resource"`
	Edge     *Edge               `json:"edge"`
	Metrics  []*Metric           `json:"metrics"`
}

// Get returns a metric associated with a name
func (t *TrafficMetrics) Get(name string) *Metric {
	for _, metric := range t.Metrics {
		if metric.Name == name {
			return metric
		}
	}

	return nil
}

// String returns a formatted string representation of this struct
func (t *TrafficMetrics) String() string {
	return fmt.Sprintf("%#v", t)
}
