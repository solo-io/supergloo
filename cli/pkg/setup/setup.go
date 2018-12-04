package setup

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"

	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/supergloo/pkg/constants"

	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

// Check if supergloo is running on the cluster and deploy it if it isn't
func InitSupergloo(opts *options.Options) error {
	// Should never happen, since InitCache gets  called first, but just in case
	if opts.Cache.KubeClient == nil {
		if err := InitCache(opts); err != nil {
			return err
		}
	}

	if !PodAppears("kube-system", opts.Cache.KubeClient, "tiller") {
		fmt.Printf("Ensuring helm is initialized on kubernetes cluster.\n")
		cmd := exec.Command("kubectl", "apply", "-f", common.HelmSetupFileName)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}
		fmt.Printf("Running helm init.\n")
		cmd = exec.Command("helm", "init", "--service-account", "tiller", "--upgrade")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}
		fmt.Printf("Waiting for Tiller pod to be ready.\n")
		if !LoopUntilPodAppears("kube-system", opts.Cache.KubeClient, "tiller") {
			return errors.Errorf("Tiller pod didn't get created")
		}
		if !LoopUntilAllPodsReadyOrTimeout("kube-system", opts.Cache.KubeClient, "tiller") {
			return errors.Errorf("Tiller pod was not ready.")
		}
		fmt.Printf("Helm is initialzed.\n")
	}

	// Supergloo needs to be installed
	if !common.Contains(opts.Cache.Namespaces, constants.SuperglooNamespace) {

		opts.Cache.KubeClient.CoreV1().Namespaces().Create(&kubecore.Namespace{
			ObjectMeta: kubemeta.ObjectMeta{
				Name: constants.SuperglooNamespace,
			},
		})

		// TODO: Deploy supergloo to kubernetes. For now, we'll assume a local server
		//fmt.Printf("Initializing supergloo on kubernetes cluster.\n")
		//cmd := exec.Command("kubectl", "apply", "-f", common.SuperglooSetupFileName)
		//cmd.Stderr = os.Stderr
		//cmd.Stdout = os.Stdout
		//if err := cmd.Run(); err != nil {
		//	return err
		//}
		//// wait for supergloo pods to be ready
		//if !LoopUntilAllPodsReadyOrTimeout(constants.SuperglooNamespace, opts.Cache.KubeClient) {
		//	return errors.Errorf("Supergloo pods did not initialize.")
		//}
		//fmt.Printf("Supergloo is ready on kubernetes cluster.\n")
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
