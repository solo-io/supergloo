package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/solo-io/solo-kit/pkg/errors"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/get/printers"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	sgConstants "github.com/solo-io/supergloo/pkg/constants"

	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"
	k8sApiExt "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SuperglooInfoClient interface {
	ListResourceTypes() []string
	ListResources(gOpts options.Get) error
}

type KubernetesInfoClient struct {
	kubeCrdClient      *k8sApiExt.CustomResourceDefinitionInterface
	meshClient         *superglooV1.MeshClient
	routingRulesClient *superglooV1.RoutingRuleClient
	resourceNameMap    map[string]string
}

func NewClient() (SuperglooInfoClient, error) {

	crdClient, err := getCrdClient()
	if err != nil {
		return nil, err
	}

	meshClient, err := common.GetMeshClient()
	if err != nil {
		return nil, err
	}

	rrClient, err := common.GetRoutingRuleClient()
	if err != nil {
		return nil, err
	}

	resourceMap, err := getResourceNames(crdClient)
	if err != nil {
		return nil, err
	}

	client := &KubernetesInfoClient{
		kubeCrdClient:      crdClient,
		meshClient:         meshClient,
		routingRulesClient: rrClient,
		resourceNameMap:    resourceMap,
	}

	return client, nil
}

// Get a list of all supergloo resources
func (client *KubernetesInfoClient) ListResourceTypes() ([]string) {
	nameSlice := make([]string, len(client.resourceNameMap))
	for k := range client.resourceNameMap {
		nameSlice = append(nameSlice, k)
	}
	return nameSlice
}

func (client *KubernetesInfoClient) ListResources(gOpts options.Get) error {
	resourceType := gOpts.Type
	resourceName := gOpts.Name
	outputFormat := gOpts.Output

	standardResourceType := client.resourceNameMap[resourceType]

	// TODO(marco): make code more generic. Ideally we don't want to enumerate the different options, but I could not
	// find an interface that all of the generated clients implement
	switch standardResourceType {
	case "mesh":
		if resourceName == "" {
			res, err := (*client.meshClient).List(sgConstants.SuperglooNamespace, clients.ListOpts{})
			if err != nil {
				return err
			}
			if outputFormat == "yaml" {
				return toYaml(res)
			}
			ri, err := FromMeshList(&res), nil
			if err != nil {
				return err
			}
			return toTable(*ri, gOpts)
		} else {
			res, err := (*client.meshClient).Read(sgConstants.SuperglooNamespace, resourceName, clients.ReadOpts{})
			if err != nil {
				return err
			}
			if outputFormat == "yaml" {
				return toYaml(res)
			}
			ri, err := FromMesh(res), nil
			if err != nil {
				return err
			}
			return toTable(*ri, gOpts)
		}
	case "routingrule":
		if resourceName == "" {
			// TODO: replace with supergloo namespace
			res, err := (*client.routingRulesClient).List("default", clients.ListOpts{})
			if err != nil {
				return err
			}
			if outputFormat == "yaml" {
				return toYaml(res)
			}
			ri, err := FromRoutingRuleList(&res), nil
			if err != nil {
				return err
			}
			return toTable(*ri, gOpts)
		} else {
			res, err := (*client.routingRulesClient).Read("default", resourceName, clients.ReadOpts{})
			if err != nil {
				return err
			}
			if outputFormat == "yaml" {
				return toYaml(res)
			}
			ri, err := FromRoutingRule(res), nil
			if err != nil {
				return err
			}
			return toTable(*ri, gOpts)
		}
	default:
		// Should not happen since we validate the resource
		return errors.Errorf(common.UnknownResourceTypeMsg, resourceType)
	}
}

// Return a client to query kubernetes CRDs
func getCrdClient() (*k8sApiExt.CustomResourceDefinitionInterface, error) {
	config, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, fmt.Errorf(common.KubeConfigError, err)
	}

	apiExtClient, err := k8sApiExt.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error building Kubernetes client: %v \n", err)
	}
	crdClient := apiExtClient.CustomResourceDefinitions()
	return &crdClient, nil
}

func toYaml(data interface{}) error {
	yml, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(yml))
	return nil
}

func toTable(resourceInfo ResourceInfo, gOpts options.Get) error {
	// Write the resource information to stdout
	writer := printers.NewTableWriter(os.Stdout)
	if err := writer.WriteLine(resourceInfo.Headers(gOpts)); err != nil {
		return err
	}
	for _, line := range resourceInfo.Resources(gOpts) {
		if err := writer.WriteLine(line); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func getResourceNames(crdClient *k8sApiExt.CustomResourceDefinitionInterface) (map[string]string, error) {
	crdList, err := (*crdClient).List(k8s.ListOptions{})
	resourceMap := make(map[string]string)
	if err != nil {
		return resourceMap,  fmt.Errorf("Error retrieving supergloo resource types. Cause: %v \n", err)
	}

	//superglooCRDs := make([]string, 0)
	for _, crd := range crdList.Items {
		if strings.Contains(crd.Name, common.SuperglooGroupName) {
			nameSpec := crd.Spec.Names
			allNames := []string{nameSpec.Singular, nameSpec.Plural}
			allNames = append(allNames, nameSpec.ShortNames...)
			for _, v := range allNames {
				resourceMap[v] = nameSpec.Singular
			}

		}
	}
	return resourceMap, nil
}
