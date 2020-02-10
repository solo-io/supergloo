package controller

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ = Describe("predicate", func() {
	var (
		statusPredicate predicate.Predicate
		ctx             context.Context
	)

	BeforeEach(func() {
		ctx = context.TODO()
		statusPredicate = IgnoreStatusUpdatePredicate(ctx)
	})

	Context("update", func() {
		It("will return true if meta has changed", func() {
			mesh1 := &v1alpha1.MeshGroup{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello": "world",
					},
				},
			}
			mesh2 := &v1alpha1.MeshGroup{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"world": "hello",
					},
				},
			}
			updateEvent := event.UpdateEvent{
				MetaOld:   mesh1,
				ObjectOld: mesh1,
				MetaNew:   mesh2,
				ObjectNew: mesh2,
			}
			Expect(statusPredicate.Update(updateEvent)).To(BeTrue())
		})

		It("will return true if spec has changed", func() {
			mesh1 := &v1alpha1.MeshGroup{
				Spec: types.MeshGroupSpec{DisplayName: "hello"},
			}
			mesh2 := &v1alpha1.MeshGroup{
				Spec: types.MeshGroupSpec{DisplayName: "world"},
			}
			updateEvent := event.UpdateEvent{
				MetaOld:   mesh1,
				ObjectOld: mesh1,
				MetaNew:   mesh2,
				ObjectNew: mesh2,
			}
			Expect(statusPredicate.Update(updateEvent)).To(BeTrue())
		})

		It("will return false if only status has changed", func() {
			mesh1 := &v1alpha1.MeshGroup{
				Status: types.MeshGroupStatus{Config: 0},
			}
			mesh2 := &v1alpha1.MeshGroup{
				Status: types.MeshGroupStatus{Config: 1},
			}
			updateEvent := event.UpdateEvent{
				MetaOld:   mesh1,
				ObjectOld: mesh1,
				MetaNew:   mesh2,
				ObjectNew: mesh2,
			}
			Expect(statusPredicate.Update(updateEvent)).To(BeFalse())
		})
	})
})
