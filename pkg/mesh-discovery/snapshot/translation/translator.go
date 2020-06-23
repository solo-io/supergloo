package translation

import (
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshots/output"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/input"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/mesh"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/meshservice"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/meshworkload"
	"github.com/solo-io/smh/pkg/mesh-discovery/utils/labelutils"
)

// the translator "reconciles the entire state of the world"
type Translator interface {
	// translates the Input Snapshot to an Output Snapshot
	Translate(in input.Snapshot) (output.Snapshot, error)
}

type translator struct {
	meshTranslator         mesh.Translator
	meshWorkloadTranslator meshworkload.Translator
	meshServiceTranslator  meshservice.Translator
}

func NewTranslator(meshTranslator mesh.Translator, meshWorkloadTranslator meshworkload.Translator, meshServiceTranslator meshservice.Translator) Translator {
	return &translator{meshTranslator: meshTranslator, meshWorkloadTranslator: meshWorkloadTranslator, meshServiceTranslator: meshServiceTranslator}
}

func (t translator) Translate(in input.Snapshot) (output.Snapshot, error) {
	meshes := t.meshTranslator.TranslateMeshes(in.Deployments())
	meshWorkloads := t.meshWorkloadTranslator.TranslateMeshWorkloads(in.Deployments(), in.DaemonSets(), in.StatefulSets(), meshes)
	meshServices := t.meshServiceTranslator.TranslateMeshServices(in.Services(), meshWorkloads)

	return output.NewLabelPartitionedSnapshot(
		labelutils.ClusterLabelKey,
		meshServices,
		meshWorkloads,
		meshes,
	)
}
