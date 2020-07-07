package translator

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/reporter"
)

// the translator "reconciles the entire state of the world"
type Translator interface {
	// translates the Input Snapshot to an Output Snapshot
	Translate(ctx context.Context, in Snapshot, reporter reporter.Reporter) (output.Snapshot, error)
}

//
//type translator struct {
//	dependencies dependencyFactory
//}
//
//func NewTranslator() Translator {
//	return &translator{dependencies: dependencyFactoryImpl{}}
//}
//
//func (t translator) Translate(ctx context.Context, in input.Snapshot) (output.Snapshot, error) {
//
//	meshTranslator := t.dependencies.makeMeshTranslator(ctx,
//		in.ConfigMaps(),
//	)
//
//	meshWorkloadTranslator := t.dependencies.makeMeshWorkloadTranslator(ctx,
//		in.Pods(),
//		in.ReplicaSets(),
//	)
//
//	meshServiceTranslator := t.dependencies.makeMeshServiceTranslator()
//
//	meshes := meshTranslator.TranslateMeshes(in.Deployments())
//
//	meshWorkloads := meshWorkloadTranslator.TranslateMeshWorkloads(
//		in.Deployments(),
//		in.DaemonSets(),
//		in.StatefulSets(),
//		meshes,
//	)
//
//	meshServices := meshServiceTranslator.TranslateMeshServices(in.Services(), meshWorkloads)
//
//	return output.NewLabelPartitionedSnapshot(
//		"mesh-discovery",
//		labelutils.ClusterLabelKey,
//		meshServices,
//		meshWorkloads,
//		meshes,
//	)
//}
//
//
//
////import (
////	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
////	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
////	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
////	"istio.io/api/networking/v1alpha3"
////)
////
////type Snapshot struct {
////
////}
////
////type plugin interface {
////	ProcessTrafficPOlciyForVirtualService(in *Snapshot, out *v1alpha3.VirtualService) error
////	ProcessTrafficPOlciyForDestinationRule(in *Snapshot, out *v1alpha3.DestinationRule) error
////}
////
////type mnt struct{
////	trafficPolicyprocessors []trafficPolicyplugin
////}
////
////func (t *mnt) TranslateMeshNetworkingSnapshot(set v1alpha1sets.MeshServiceSet) v1alpha3sets.DestinationRuleSet {
////	drset := v1alpha3sets.NewDestinationRuleSet()
////	for _, meshService := range set.List() {
////		dr := makeDrForSvc(meshService)
////
////		for _, processor := range t.processors {
////			processor.ProcessDestinationRule(in, dr)
////		}
////
////		drset.Insert(dr)
////	}
////
////	return drset
////}
