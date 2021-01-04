package virtualrouter

import (
	"context"
	"fmt"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/appmesh/traffictarget/internal/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
)

type Translator interface {
	Translate(
		ctx context.Context,
		localSnapshot input.LocalSnapshot,
		remoteSnapshot input.RemoteSnapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		reporter reporting.Reporter,
	) *appmeshv1beta2.VirtualRouter
}

type translator struct{}

func NewVirtualRouterTranslator() Translator {
	return &translator{}
}

func (t *translator) Translate(
	ctx context.Context,
	localSnapshot input.LocalSnapshot,
	remoteSnapshot input.RemoteSnapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	reporter reporting.Reporter,
) *appmeshv1beta2.VirtualRouter {

	kubeService := trafficTarget.Spec.GetKubeService()
	if kubeService == nil {
		// TODO non kube services currently unsupported
		return nil
	}

	routeTranslator := newRouteTranslator(reporter, localSnapshot.DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets(), localSnapshot.DiscoveryMeshGlooSoloIov1Alpha2Workloads())
	routes := routeTranslator.getRoutes(trafficTarget)
	if len(routes) == 0 {
		// There are no routes, so we don't need to create a virtual router
		return nil
	}

	listenerTranslator := newListenerTranslator()
	listeners := listenerTranslator.getListeners(trafficTarget)

	meshRef, err := utils.GetAppMeshMeshRef(trafficTarget, localSnapshot.DiscoveryMeshGlooSoloIov1Alpha2Meshes())
	if err != nil {
		// TODO joekelley
		return nil
	}

	meta := metautils.TranslatedObjectMeta(
		trafficTarget.Spec.GetKubeService().Ref,
		trafficTarget.Annotations,
	)

	// If the router exists, we must set the mesh ref on the virtual router.
	// If it does not exist, we must not set the mesh ref on the virtual router,
	// as the app mesh controller's admission controller will block the write.
	existingRouter, _ := remoteSnapshot.AppmeshK8SAwsv1Beta2VirtualRouters().Find(&meta)
	if existingRouter == nil {
		meshRef = nil
	}

	// This is the default name written back by the AWS controller.
	// We must provide it explicitly, else the App Mesh controller's
	// validating admission webhook will reject our changes on update.
	awsName := fmt.Sprintf("%s_%s", meta.Name, meta.Namespace)

	return &appmeshv1beta2.VirtualRouter{
		ObjectMeta: meta,
		Spec: appmeshv1beta2.VirtualRouterSpec{
			AWSName:   &awsName,
			MeshRef:   meshRef,
			Listeners: listeners,
			Routes:    routes,
		},
	}
}
