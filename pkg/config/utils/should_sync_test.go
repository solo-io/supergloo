package utils_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/api/external/kubernetes/customresourcedefinition"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	. "github.com/solo-io/supergloo/pkg/config/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ShouldSync", func() {
	It("returns true when all the required crds are present in the list", func() {
		crdNames := []string{"a", "b", "c"}

		Expect(ShouldSync(context.TODO(), "", crdNames, crds(crdNames...))).To(BeTrue())
		Expect(ShouldSync(context.TODO(), "", crdNames, crds("a", "b"))).To(BeFalse())
	})
})

func crds(names ...string) kubernetes.CustomResourceDefinitionList {
	var crdList kubernetes.CustomResourceDefinitionList
	for _, name := range names {
		crdList = append(crdList, &kubernetes.CustomResourceDefinition{
			CustomResourceDefinition: customresourcedefinition.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{Name: name},
			}})
	}
	return crdList
}
