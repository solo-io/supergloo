package vm_validation_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	vm_validation "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("mesh ref finder", func() {
	var (
		ctrl          *gomock.Controller
		ctx           context.Context
		meshClient    *mock_zephyr_discovery.MockMeshClient
		meshRefFinder vm_validation.VirtualMeshFinder

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		meshClient = mock_zephyr_discovery.NewMockMeshClient(ctrl)
		meshRefFinder = vm_validation.NewVitualMeshFinder(meshClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will fail if initial mesh list fails", func() {
		meshClient.EXPECT().
			List(ctx).
			Return(nil, testErr)

		_, err := meshRefFinder.GetMeshesForVirtualMesh(ctx, &networking_v1alpha1.VirtualMesh{})
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(testErr))
	})

	It("will reutrn nil, nil with no refs as input", func() {
		meshClient.EXPECT().
			List(ctx).
			Return(nil, nil)
		list, err := meshRefFinder.GetMeshesForVirtualMesh(ctx, &networking_v1alpha1.VirtualMesh{})
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(BeNil())
	})

	It("will return an error containing all invalid refs", func() {
		meshList := &discovery_v1alpha1.MeshList{}
		refs := []*core_types.ResourceRef{
			{
				Name:      "name1",
				Namespace: "namespace1",
			},
			{
				Name:      "name2",
				Namespace: "namespace2",
			},
		}
		vm := &networking_v1alpha1.VirtualMesh{
			Spec: types.VirtualMeshSpec{
				Meshes: refs,
			},
		}
		meshClient.EXPECT().
			List(ctx).
			Return(meshList, nil)
		_, err := meshRefFinder.GetMeshesForVirtualMesh(ctx, vm)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(vm_validation.InvalidMeshRefsError([]string{
			fmt.Sprintf("%s.%s", refs[0].GetName(), refs[0].GetNamespace()),
			fmt.Sprintf("%s.%s", refs[1].GetName(), refs[1].GetNamespace()),
		})))
	})

	It("will return an error containing all invalid refs", func() {
		meshList := &discovery_v1alpha1.MeshList{
			Items: []discovery_v1alpha1.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "name1",
						Namespace: "namespace1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "name2",
						Namespace: "namespace2",
					},
				},
			},
		}
		refs := []*core_types.ResourceRef{
			{
				Name:      "name1",
				Namespace: "namespace1",
			},
			{
				Name:      "name2",
				Namespace: "namespace2",
			},
		}
		vm := &networking_v1alpha1.VirtualMesh{
			Spec: types.VirtualMeshSpec{
				Meshes: refs,
			},
		}
		meshClient.EXPECT().
			List(ctx).
			Return(meshList, nil)
		list, err := meshRefFinder.GetMeshesForVirtualMesh(ctx, vm)
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(ConsistOf(&meshList.Items[0], &meshList.Items[1]))
	})
})
