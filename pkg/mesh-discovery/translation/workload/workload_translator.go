package workload

import (
	"context"

	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/types"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./workload_translator.go -destination mocks/workload_translator.go

// the mesh-workload translator converts deployments with injected sidecars into Workload CRs
type Translator interface {
	TranslateWorkloads(deployments appsv1sets.DeploymentSet, daemonSets appsv1sets.DaemonSetSet, statefulSets appsv1sets.StatefulSetSet, meshes v1alpha2sets.MeshSet) v1alpha2sets.WorkloadSet
}

type translator struct {
	ctx              context.Context
	workloadDetector detector.WorkloadDetector
}

func NewTranslator(
	ctx context.Context,
	workloadDetector detector.WorkloadDetector,
) Translator {
	return &translator{ctx: ctx, workloadDetector: workloadDetector}
}

func (t *translator) TranslateWorkloads(deployments appsv1sets.DeploymentSet, daemonSets appsv1sets.DaemonSetSet, statefulSets appsv1sets.StatefulSetSet, meshes v1alpha2sets.MeshSet) v1alpha2sets.WorkloadSet {
	var workloads []types.Workload
	for _, deployment := range deployments.List() {
		workloads = append(workloads, types.ToWorkload(deployment))
	}
	for _, daemonSet := range daemonSets.List() {
		workloads = append(workloads, types.ToWorkload(daemonSet))
	}
	for _, statefulSet := range statefulSets.List() {
		workloads = append(workloads, types.ToWorkload(statefulSet))
	}

	workloadSet := v1alpha2sets.NewWorkloadSet()

	for _, workload := range workloads {
		discoveredWorkload := t.workloadDetector.DetectWorkload(workload, meshes)
		if discoveredWorkload == nil {
			continue
		}
		contextutils.LoggerFrom(t.ctx).Debugf("detected workload %v", sets.Key(discoveredWorkload))
		workloadSet.Insert(discoveredWorkload)
	}
	return workloadSet
}
