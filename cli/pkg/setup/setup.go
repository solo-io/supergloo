package setup

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"

	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/solo-io/supergloo/pkg/constants"
	k8sV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitCache(opts *options.Options) error {
	// Get a kube client
	kube, err := common.GetKubernetesClient()
	if err != nil {
		return err
	}
	opts.Cache.KubeClient = kube

	// Get all namespaces
	list, err := kube.CoreV1().Namespaces().List(v1.ListOptions{IncludeUninitialized: false})
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
	secretClient, err := common.GetSecretClient()
	if err != nil {
		return err
	}
	//   2. get client resources for each namespace
	// 2.a secrets, prime the map
	opts.Cache.NsResources = make(map[string]*options.NsResource)
	for _, ns := range namespaces {
		secretList, err := (*secretClient).List(ns, clients.ListOpts{})
		if err != nil {
			return err
		}
		var secrets = []string{}
		for _, m := range secretList {
			secrets = append(secrets, m.Metadata.Name)
		}

		// prime meshes
		var meshes = []string{}
		opts.Cache.NsResources[ns] = &options.NsResource{
			Meshes:  meshes,
			Secrets: secrets,
		}
	}
	// 2.b meshes
	// meshes are categorized by their installation namespace, which may be different than the mesh CRD's namespace
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
			opts.Cache.NsResources[iNs].Meshes = append(opts.Cache.NsResources[iNs].Meshes, m.Metadata.Name)
		}

	}

	return nil
}

// Check if  supergloo is running on the cluster and deploy it if it isn't
func Init(opts *options.Options) error {

	tempDir, err := ioutil.TempDir("", "supergloo")
	defer os.RemoveAll(tempDir)

	// Should never happen, since InitCache gets  called first, but just in case
	if opts.Cache.KubeClient == nil {
		if err := InitCache(opts); err != nil {
			return err
		}
	}

	// Create supergloo namespace if it does not exist
	if !common.Contains(opts.Cache.Namespaces, constants.SuperglooNamespace) {
		opts.Cache.KubeClient.CoreV1().Namespaces().Create(
			&k8sV1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: constants.SuperglooNamespace,
				},
			})
	}

	// Get the pods in the supergloo namespace
	pods, err := opts.Cache.KubeClient.CoreV1().Pods(constants.SuperglooNamespace).List(v1.ListOptions{})
	if err != nil {
		return err
	}

	// Very hacky way of determining if supergloo is running
	var installed bool
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "supergloo") {
			installed = true
		}
	}

	if !installed {

		absPathToFile, err := downloadFile(tempDir)
		if err != nil {
			return err
		}

		cmd := exec.Command("kubectl", "apply", "-f", absPathToFile)
		return cmd.Run() //TODO: when this returns the resources might still be pending
	}

	return nil
}

func downloadFile(dir string) (string, error) {

	// Create the file
	absFilePath := filepath.Join(dir, common.SuperglooSetupFileName)
	out, err := os.Create(absFilePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(common.SuperglooResourceDefinitionUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return absFilePath, nil
}
