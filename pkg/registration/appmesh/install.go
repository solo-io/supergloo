package appmesh

import "io"

//go:generate mockgen -destination=./install_mock.go -source install.go -package appmesh

// Hide the client in k8s.io/helm/pkg/kube behind our own interface so we can mock it.
type Installer interface {
	// Delete deletes Kubernetes resources from an io.reader.
	Delete(namespace string, reader io.Reader) error
	// Create creates Kubernetes resources from an io.reader.
	Create(namespace string, reader io.Reader, timeout int64, shouldWait bool) error
}
