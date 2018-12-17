package kube_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/supergloo/pkg/kube"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

// source: https://raw.githubusercontent.com/linkerd/linkerd2/master/cli/install/template.go
const linkerdCrdYaml = `### Service Profile CRD ###
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: serviceprofiles.linkerd.io
spec:
  group: linkerd.io
  version: v1alpha1
  scope: Namespaced
  names:
    plural: serviceprofiles
    singular: serviceprofile
    kind: ServiceProfile 
    shortNames:
    - sp
`

var _ = Describe("Crd", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		testCrds  []*v1beta1.CustomResourceDefinition
		apiExts   apiexts.Interface
		crdClient kube.CrdClient
	)
	BeforeEach(func() {
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		apiExts, err = apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		testCrds, err = kube.CrdsFromManifest(linkerdCrdYaml)
		Expect(err).NotTo(HaveOccurred())
		crdClient = kube.NewKubeCrdClient(apiExts)
	})
	AfterEach(func() {
		var crdsToDelete []string
		for _, crd := range testCrds {
			crdsToDelete = append(crdsToDelete, crd.Name)
		}
		crdClient.DeleteCrds(crdsToDelete...)
	})
	It("creates crds", func() {
		crdClient.CreateCrds(testCrds...)
		crdList, err := apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, testCrd := range testCrds {
			var found bool
			for _, actual := range crdList.Items {
				if testCrd.Name == actual.Name {
					found = true
					// set by apiserver
					testCrd.Spec.Names.ListKind = testCrd.Spec.Names.Kind + "List"
					Expect(testCrd.Spec).To(Equal(actual.Spec))
					break
				}
			}
			Expect(found).To(BeTrue())
		}
	})
})
