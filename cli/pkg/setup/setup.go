package setup

import (
	"strings"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"

	"k8s.io/client-go/kubernetes"

	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Checks whether the cluster that the kubeconfig points at is available
// The timeout for the kubernetes client is set to a low value to notify the user of the failure
func CheckConnection() error {
	config, err := common.GetKubernetesConfig(time.Second)
	if err != nil {
		return err
	}
	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	_, err = kube.CoreV1().Namespaces().List(kubemeta.ListOptions{})

	return err
}

func InitCache(opts *options.Options) error {
	// Get a kube client
	kube, err := common.GetKubernetesClient()
	if err != nil {
		return err
	}
	opts.Cache.KubeClient = kube

	// Get all namespaces
	list, err := kube.CoreV1().Namespaces().List(kubemeta.ListOptions{IncludeUninitialized: false})
	if err != nil {
		return err
	}
	var namespaces []string
	for _, ns := range list.Items {
		namespaces = append(namespaces, ns.ObjectMeta.Name)
	}
	opts.Cache.Namespaces = namespaces

	// Get key resources by ns
	//   1. gather clients
	meshClient, err := common.GetMeshClient()
	if err != nil {
		return err
	}
	istioSecretClient, err := common.GetIstioSecretClient()
	if err != nil {
		return err
	}
	glooSecretClient, err := common.GetGlooSecretClient()
	if err != nil {
		return err
	}
	upstreamClient, err := common.GetUpstreamClient()
	if err != nil {
		return err
	}
	//   2. get client resources for each namespace
	// 2.a secrets, meshes, prime the mesh-by-installation-ns map
	opts.Cache.NsResources = make(map[string]*options.NsResource)
	for _, ns := range namespaces {
		meshList, err := (*meshClient).List(ns, clients.ListOpts{})
		if err != nil {
			return err
		}
		var meshes = []string{}
		for _, m := range meshList {
			meshes = append(meshes, m.Metadata.Name)
		}
		istioSecretList, err := (*istioSecretClient).List(ns, clients.ListOpts{})
		if err != nil {
			return err
		}
		var istioSecrets = []string{}
		for _, m := range istioSecretList {
			istioSecrets = append(istioSecrets, m.Metadata.Name)
		}
		glooSecretList, err := (*glooSecretClient).List(ns, clients.ListOpts{})
		if err != nil {
			return err
		}
		var glooSecrets = []string{}
		for _, m := range glooSecretList {
			glooSecrets = append(glooSecrets, m.Metadata.Name)
		}
		upstreamList, err := (*upstreamClient).List(ns, clients.ListOpts{})
		if err != nil {
			return err
		}
		var upstreams = []string{}
		for _, m := range upstreamList {
			upstreams = append(upstreams, m.Metadata.Name)
		}

		// prime meshes
		var meshesByInstallNs = []core.ResourceRef{}
		opts.Cache.NsResources[ns] = &options.NsResource{
			MeshesByInstallNs: meshesByInstallNs,
			Meshes:            meshes,
			IstioSecrets:      istioSecrets,
			GlooSecrets:       glooSecrets,
			Upstreams:         upstreams,
		}
	}
	// 2.c meshes by installation namespace
	// meshes are also categorized by their installation namespace, which may be different than the mesh CRD's namespace
	for _, ns := range namespaces {
		meshList, err := (*meshClient).List(ns, clients.ListOpts{})
		if err != nil {
			return err
		}
		for _, m := range meshList {
			var iNs string
			// dial in by resource type
			switch spec := m.MeshType.(type) {
			case *superglooV1.Mesh_Consul:
				iNs = spec.Consul.InstallationNamespace
			case *superglooV1.Mesh_Linkerd2:
				iNs = spec.Linkerd2.InstallationNamespace
			case *superglooV1.Mesh_Istio:
				iNs = spec.Istio.InstallationNamespace
			}
			if iNs != "" {
				opts.Cache.NsResources[iNs].MeshesByInstallNs = append(
					opts.Cache.NsResources[iNs].MeshesByInstallNs,
					core.ResourceRef{
						Name:      m.Metadata.Name,
						Namespace: m.Metadata.Namespace,
					})
			}
		}
	}

	return nil
}

func PodAppears(namespace string, client *kubernetes.Clientset, podName string) bool {
	podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
	if err != nil {
		return false
	}
	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, podName) {
			return true
		}
	}
	return false
}

func LoopUntilPodAppears(namespace string, client *kubernetes.Clientset, podName string) bool {
	for i := 0; i < 30; i++ {
		if PodAppears(namespace, client, podName) {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func AllPodsReadyOrSucceeded(namespace string, client *kubernetes.Clientset, podNames ...string) bool {
	podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
	if err != nil {
		return false
	}
	done := true
	for _, pod := range podList.Items {
		if len(podNames) > 0 && !common.ContainsSubstring(podNames, pod.Name) {
			continue
		}
		for _, condition := range pod.Status.Conditions {
			if pod.Status.Phase == kubecore.PodSucceeded {
				continue
			}
			if condition.Type == kubecore.PodReady && condition.Status != kubecore.ConditionTrue {
				done = false
			}
		}
	}
	return done
}

func LoopUntilAllPodsReadyOrTimeout(namespace string, client *kubernetes.Clientset, podNames ...string) bool {
	for i := 0; i < 30; i++ {
		if AllPodsReadyOrSucceeded(namespace, client, podNames...) {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}
