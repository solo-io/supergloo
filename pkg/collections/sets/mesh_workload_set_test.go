package sets_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/collections/sets"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MeshWorkloadSet", func() {
	var (
		setA         sets.MeshServiceSet
		setB         sets.MeshServiceSet
		meshServiceA *v1alpha1.MeshService
		meshServiceB *v1alpha1.MeshService
		meshServiceC *v1alpha1.MeshService
	)

	BeforeEach(func() {
		setA = sets.NewMeshServiceSet()
		setB = sets.NewMeshServiceSet()
		meshServiceA = &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{Name: "nameA", Namespace: "nsA"},
		}
		meshServiceB = &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{Name: "nameB", Namespace: "nsB"},
		}
		meshServiceC = &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{Name: "nameC", Namespace: "nsC"},
		}
	})

	It("should insert", func() {
		setA.Insert(meshServiceA)
		list := setA.List()
		Expect(list).To(ConsistOf(meshServiceA))
		setA.Insert(meshServiceB, meshServiceC)
		list = setA.List()
		Expect(list).To(ConsistOf(meshServiceA, meshServiceB, meshServiceC))
	})

	It("should return set existence", func() {
		setA.Insert(meshServiceA)
		Expect(setA.Has(meshServiceA)).To(BeTrue())
		Expect(setA.Has(meshServiceB)).To(BeFalse())
		setA.Insert(meshServiceB, meshServiceC)
		Expect(setA.Has(meshServiceA)).To(BeTrue())
		Expect(setA.Has(meshServiceB)).To(BeTrue())
		Expect(setA.Has(meshServiceC)).To(BeTrue())
	})

	It("should return set equality", func() {
		setB.Insert(meshServiceA, meshServiceB, meshServiceC)
		setA.Insert(meshServiceA)
		Expect(setA.Equal(setB)).To(BeFalse())
		setA.Insert(meshServiceC, meshServiceB)
		Expect(setA.Equal(setB)).To(BeTrue())
	})

	It("should delete", func() {
		setA.Insert(meshServiceA, meshServiceB, meshServiceC)
		Expect(setA.Has(meshServiceA)).To(BeTrue())
		setA.Delete(meshServiceA)
		Expect(setA.Has(meshServiceA)).To(BeFalse())
	})

	It("should union two sets and return new set", func() {
		setA.Insert(meshServiceA, meshServiceB)
		setB.Insert(meshServiceA, meshServiceB, meshServiceC)
		unionSet := setA.Union(setB)
		Expect(unionSet.List()).To(ConsistOf(meshServiceA, meshServiceB, meshServiceC))
		Expect(unionSet).ToNot(BeIdenticalTo(setA))
		Expect(unionSet).ToNot(BeIdenticalTo(setB))
	})

	It("should take the difference of two sets and return new set", func() {
		setA.Insert(meshServiceA, meshServiceB)
		setB.Insert(meshServiceA, meshServiceB, meshServiceC)
		differenceA := setA.Difference(setB)
		Expect(differenceA.List()).To(BeEmpty())
		Expect(differenceA.Map()).To(BeEmpty())
		Expect(differenceA).ToNot(BeIdenticalTo(setA))

		differenceB := setB.Difference(setA)
		Expect(differenceB.List()).To(ConsistOf(meshServiceC))
		Expect(differenceB.Map()).To(HaveKeyWithValue(clients.ToUniqueSingleClusterString(meshServiceC.ObjectMeta), meshServiceC))
		Expect(differenceB).ToNot(BeIdenticalTo(setB))
	})

	It("should take the intersection of two sets and return new set", func() {
		setA.Insert(meshServiceA, meshServiceB)
		setB.Insert(meshServiceA, meshServiceB, meshServiceC)
		intersectionA := setA.Intersection(setB)
		Expect(intersectionA.List()).To(ConsistOf(meshServiceA, meshServiceB))
		Expect(intersectionA.Map()).To(HaveKeyWithValue(clients.ToUniqueSingleClusterString(meshServiceA.ObjectMeta), meshServiceA))
		Expect(intersectionA.Map()).To(HaveKeyWithValue(clients.ToUniqueSingleClusterString(meshServiceB.ObjectMeta), meshServiceB))
		Expect(intersectionA).ToNot(BeIdenticalTo(setA))
	})
})
