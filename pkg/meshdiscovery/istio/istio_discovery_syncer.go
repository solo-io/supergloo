package istio

import (
	"context"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type istioDiscoverySyncer struct {
	writeNamespace   string
	meshReconciler   v1.MeshReconciler
	meshPolicyClient v1alpha1.MeshPolicyClient
	apiexts          apiexts.Interface
}

func NewIstioDiscoverySyncer(meshReconciler v1.MeshReconciler) v1.DiscoverySyncer {
	return &istioDiscoverySyncer{meshReconciler: meshReconciler}
}

func (i *istioDiscoverySyncer) Sync(ctx context.Context, snap *v1.DiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, "istio-mesh-discovery")
}

type pilotDeployment struct {
	version, namespace string
}

func (i *istioDiscoverySyncer) desiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	meshPolicyCrdRegistered, err := detectMeshPolicyCrd(i.apiexts)
	if err != nil {
		return nil, err
	}
	if !meshPolicyCrdRegistered {
		return nil, nil
	}

	pilots, err := detectPilotDeployments(snap.Deployments)
	if err != nil {
		return nil, err
	}
	if len(pilots) == 0 {
		return nil, nil
	}

	// https://istio.io/docs/tasks/security/authn-policy/#globally-enabling-istio-mutual-tls
	defaultMeshPolicy, err := i.meshPolicyClient.Read("default", clients.ReadOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	globalMtlsEnabled := func() bool {
		for _, peer := range defaultMeshPolicy.GetPeers() {
			return peer.GetMtls() != nil
		}
		return false
	}()

	var istioMeshes v1.MeshList
	for _, pilot := range pilots {
		var autoInjectionEnabled bool
		sidecarInjector, err := snap.Deployments.Find(pilot.namespace, "istio-sidecar-injector")
		if err == nil && (sidecarInjector.Spec.Replicas == nil || *sidecarInjector.Spec.Replicas > 0) {
			autoInjectionEnabled = true
		}

		var smiEnabled bool
		smiAdapter, err := snap.Deployments.Find(pilot.namespace, "smi-adapter-istio")
		if err == nil && (smiAdapter.Spec.Replicas == nil || *smiAdapter.Spec.Replicas > 0) {
			smiEnabled = true
		}

		// https://istio.io/docs/tasks/security/plugin-ca-cert/#plugging-in-the-existing-certificate-and-key
		var rootCa *core.ResourceRef
		customRootCa, err := snap.Tlssecrets.Find(pilot.namespace, "cacerts")
		if err == nil {
			root := customRootCa.Metadata.Ref()
			rootCa = &root
		}


		mtlsConfig := &v1.MtlsConfig{
			MtlsEnabled:     globalMtlsEnabled,
			RootCertificate: rootCa,
		}

		meshUpstreams, err := detectMeshUpstreams(pilot.namespace, snap.Pods, snap.Upstreams)
		if err != nil {
			return nil, err
		}

		istioMesh := &v1.Mesh{
			Metadata: core.Metadata{
				Name:      pilot.namespace + "-istio",
				Namespace: i.writeNamespace,
			},
			MeshType: &v1.Mesh_Istio{
				Istio: &v1.IstioMesh{
					InstallationNamespace: pilot.namespace,
					Version:               pilot.version,
				},
			},
			MtlsConfig: mtlsConfig,
			SmiEnabled: smiEnabled,
			DiscoveryMetadata: &v1.DiscoveryMetadata{
				EnableAutoInject: autoInjectionEnabled,
				MtlsConfig:       mtlsConfig,
				Upstreams:        meshUpstreams,
			},
		}
	}

}

func detectPilotDeployments(deployments kubernetes.DeploymentList) ([]pilotDeployment, error) {
	var pilots []pilotDeployment
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if strings.Contains(container.Image, "istio/pilot") {
				split := strings.Split(container.Image, ":")
				if len(split) != 2 {
					return nil, errors.Errorf("invalid or unexpected image format for pilot: %v", container.Image)
				}
				pilots = append(pilots, pilotDeployment{version: split[1], namespace: deployment.Namespace})
			}
		}
	}
	return pilots, nil
}

func detectMeshPolicyCrd(apiexts apiexts.Interface) (bool, error) {
	_, err := apiexts.ApiextensionsV1beta1().CustomResourceDefinitions().Get(v1alpha1.MeshPolicyCrd.FullName(), metav1.GetOptions{})
	if err == nil {
		return true, nil
	}
	if kubeerrs.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

func detectMeshUpstreams(pilotNamespace string, pods kubernetes.PodList, upstreams gloov1.UpstreamList) ([]*core.ResourceRef, error) {

}