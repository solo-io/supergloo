package metrics

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

// Direction defines whether an edge is sending traffic to a resource or
// receiving traffic from a resource.
type Direction string

const (
	// To is used when an edge is sending traffic to a resource
	To Direction = "to"
	// From is used when an edge is receiving traffic from a resource
	From Direction = "from"
)

// Edge describes the other resource that metrics are associated with
type Edge struct {
	Direction Direction           `json:"direction"`
	Resource  *v1.ObjectReference `json:"resource"`
}

// String returns a formatted string representation of this struct
func (e *Edge) String() string {
	return fmt.Sprintf("%#v", e)
}
