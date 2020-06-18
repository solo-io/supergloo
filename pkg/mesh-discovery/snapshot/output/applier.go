package output

import "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshots/output"

// the Applier applies the output snapshot to the backing persistence store (i.e. Kube CRDs)
type Applier interface {
	Apply(output output.Snapshot) error
}
