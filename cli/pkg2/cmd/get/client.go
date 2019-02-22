package get

import (
	"fmt"
	"strings"

	"github.com/solo-io/solo-kit/pkg/errors"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/get/printers"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	superglooV1 "github.com/solo-io/supergloo/pkg2/api/v1"
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

	crdClient, err := common.GetKubeCrdClient()
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
func (client *KubernetesInfoClient) ListResourceTypes() []string {
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
	namespace := gOpts.Namespace

	standardResourceType := client.resourceNameMap[resourceType]

	switch standardResourceType {
	case "mesh":
		mList := superglooV1.MeshList{}
		if resourceName == "" {
			res, err := (*client.meshClient).List(namespace, clients.ListOpts{})
			if err != nil {
				return err
			}
			mList = append(mList, res...)
		} else {
			res, err := (*client.meshClient).Read(namespace, resourceName, clients.ReadOpts{})
			if err != nil {
				return err
			}
			mList = append(mList, res)
		}
		return printers.MeshTable(&mList, outputFormat, "")
	case "routingrule":
		rrList := superglooV1.RoutingRuleList{}
		if resourceName == "" {
			res, err := (*client.routingRulesClient).List(namespace, clients.ListOpts{})
			if err != nil {
				return err
			}
			rrList = append(rrList, res...)
		} else {
			res, err := (*client.routingRulesClient).Read(namespace, resourceName, clients.ReadOpts{})
			if err != nil {
				return err
			}
			rrList = append(rrList, res)
		}
		return printers.RoutingRuleTable(rrList, outputFormat, "")
	default:
		// Should not happen since we validate the resource
		return errors.Errorf(common.UnknownResourceTypeMsg, resourceType)
	}
}

func getResourceNames(crdClient *k8sApiExt.CustomResourceDefinitionInterface) (map[string]string, error) {
	crdList, err := (*crdClient).List(k8s.ListOptions{})
	resourceMap := make(map[string]string)
	if err != nil {
		return resourceMap, fmt.Errorf("Error retrieving supergloo resource types. Cause: %v \n", err)
	}

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
