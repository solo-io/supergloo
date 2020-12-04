package virtualservice

import (
	"context"
	"fmt"

	"github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/appmesh/traffictarget/internal/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Translator interface {
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		virtualRouter *appmeshv1beta2.VirtualRouter,
		reporter reporting.Reporter,
	) []*appmeshv1beta2.VirtualService
}

type translator struct{}

func NewVirtualServiceTranslator() Translator {
	return &translator{}
}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	virtualRouter *appmeshv1beta2.VirtualRouter,
	reporter reporting.Reporter,
) []*appmeshv1beta2.VirtualService {

	kubeService := trafficTarget.Spec.GetKubeService()
	if kubeService == nil {
		// TODO non kube services currently unsupported
		return nil
	}

	backingWorkloads := workloadutils.FindBackingWorkloads(trafficTarget.Spec.GetKubeService(), in.Workloads())
	if len(backingWorkloads) == 0 {
		contextutils.LoggerFrom(ctx).Warnf("Found no backing workloads for traffic target %s", sets.Key(trafficTarget))
		return nil
	}

	// Create a virtual service for all backing workloads.
	// We do this in order to have a virtual service for every virtual node
	// in the event that no router is provided, such that virtual services
	// are still available to clients after a traffic policy is deleted.
	// TODO joekelley think this thru
	// TODO joekelley remove special ref
	virtualServices := make([]*appmeshv1beta2.VirtualService, 0, len(backingWorkloads))
	for _, workload := range backingWorkloads {
		arn := workload.Spec.AppMesh.VirtualNodeArn
		meta := metautils.TranslatedObjectMeta(
			myRef(trafficTarget.Spec.GetKubeService().Ref),
			trafficTarget.Annotations,
		)

		awsMeshRef, err := utils.GetAppMeshMeshRef(trafficTarget, in.Meshes())
		if err != nil {
			// TODO joekelley
		}

		virtualServices = append(virtualServices, getVirtualService(meta, virtualRouter, awsMeshRef, arn))
	}

	return virtualServices
}

func getVirtualService(
	meta metav1.ObjectMeta,
	virtualRouter *appmeshv1beta2.VirtualRouter,
	meshRef *appmeshv1beta2.MeshReference,
	arn string,
) *appmeshv1beta2.VirtualService {
	var provider *appmeshv1beta2.VirtualServiceProvider
	if virtualRouter != nil {
		provider = &appmeshv1beta2.VirtualServiceProvider{
			VirtualRouter: &appmeshv1beta2.VirtualRouterServiceProvider{
				VirtualRouterRef: &appmeshv1beta2.VirtualRouterReference{
					Namespace: &meta.Namespace,
					Name:      meta.Name,
				},
			},
		}
	} else {
		provider = &appmeshv1beta2.VirtualServiceProvider{
			VirtualNode: &appmeshv1beta2.VirtualNodeServiceProvider{
				VirtualNodeARN: &arn,
			},
		}
	}

	// This is the default name written back by the AWS controller.
	// We must provide it explicitly, else the App Mesh controller's
	// validating admission webhook will reject our changes on update.
	awsName := fmt.Sprintf("%s.%s", meta.Name, meta.Namespace)
	return &appmeshv1beta2.VirtualService{
		ObjectMeta: meta,
		Spec: v1beta2.VirtualServiceSpec{
			MeshRef:  meshRef,
			AWSName:  &awsName,
			Provider: provider,
		},
	}
}

func myRef(ref ezkube.ClusterResourceId) ezkube.ClusterResourceId {
	output := &v1.ClusterObjectRef{
		Name:        ref.GetName() + "-test",
		Namespace:   ref.GetNamespace(),
		ClusterName: ref.GetClusterName(),
	}
	return output
}
