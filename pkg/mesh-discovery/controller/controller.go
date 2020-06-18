package controller

import "sigs.k8s.io/controller-runtime/pkg/manager"

// the mesh-discovery controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
type Controller interface {
	manager.Runnable
}
