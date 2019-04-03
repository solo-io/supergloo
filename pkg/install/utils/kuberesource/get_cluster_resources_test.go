package kuberesource

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/supergloo/test/e2e/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("GetClusterResources", func() {
	var ns string
	labelSetA := map[string]string{"a": "b"}
	var cm1, cm2, cm3 v1.ConfigMap
	BeforeEach(func() {
		ns = "test" + testutils.RandString(4)
		cm1, cm2, cm3 = utils.ConfigMap(ns, "a1", "data1", labelSetA),
			utils.ConfigMap(ns, "a2", "data2", labelSetA),
			utils.ConfigMap(ns, "a3", "data3", labelSetA)
		utils.MustCreateNs(ns)
		utils.MustCreateConfigMap(cm1)
		utils.MustCreateConfigMap(cm2)
		utils.MustCreateConfigMap(cm3)
	})
	AfterEach(func() {
		utils.MustDeleteNs(ns)
	})
	It("gets all resources in the cluster, period", func() {
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		allRes, err := GetClusterResources(context.TODO(), cfg, func(resource schema.GroupVersionResource) bool {
			// just get configmaps
			if resource.Resource != "configmaps" {
				return true
			}
			return false
		})
		Expect(err).NotTo(HaveOccurred())
		cmInOurNs := allRes.Filter(func(resource *unstructured.Unstructured) bool {
			return resource.GetNamespace() != ns
		})

		expected1, err := ConvertToUnstructured(&cm1)
		Expect(err).NotTo(HaveOccurred())
		expected2, err := ConvertToUnstructured(&cm2)
		Expect(err).NotTo(HaveOccurred())
		expected3, err := ConvertToUnstructured(&cm3)
		Expect(err).NotTo(HaveOccurred())
		expected := UnstructuredResources{expected1, expected2, expected3}

		Expect(cmInOurNs).To(HaveLen(3))
		for i := range expected {
			actual := cmInOurNs[i]
			delete(actual.Object, "metadata")
			delete(expected[i].Object, "metadata")
			Expect(actual).To(Equal(expected[i]))
		}
	})
})
