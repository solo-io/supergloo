package internal_watcher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	. "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/secret-event-handler/internal"
	k8s_core_types "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("common", func() {
	Context("hasRequiredMetadata", func() {
		var meta metav1.Object
		obj := &k8s_core_types.Secret{}
		It("will not work with just labels", func() {
			meta = &metav1.ObjectMeta{
				Labels: map[string]string{mc_manager.MultiClusterLabel: "true"},
			}
			Expect(HasRequiredMetadata(meta, obj)).To(BeFalse())
		})
		It("will not work with just namespace", func() {
			meta = &metav1.ObjectMeta{
				Namespace: container_runtime.GetWriteNamespace(),
			}
			Expect(HasRequiredMetadata(meta, obj)).To(BeFalse())
		})
		It("will not work with one condition", func() {
			meta = &metav1.ObjectMeta{
				Labels:    map[string]string{mc_manager.MultiClusterLabel: "true"},
				Namespace: container_runtime.GetWriteNamespace(),
			}
			Expect(HasRequiredMetadata(meta, obj)).To(BeTrue())
		})
	})
	Context("will fire an update event if either new or old matches", func() {
		var (
			updateEvent event.UpdateEvent
			pred        *MultiClusterPredicate

			matchingMeta = &metav1.ObjectMeta{
				Labels:    map[string]string{mc_manager.MultiClusterLabel: "true"},
				Namespace: container_runtime.GetWriteNamespace(),
			}

			badMeta = &metav1.ObjectMeta{
				Labels: map[string]string{mc_manager.MultiClusterLabel: "true"},
			}
		)

		BeforeEach(func() {
			updateEvent = event.UpdateEvent{}
			pred = &MultiClusterPredicate{}
		})
		It("matches old object", func() {
			updateEvent.MetaOld = matchingMeta
			updateEvent.MetaNew = badMeta
			Expect(pred.Update(updateEvent)).To(BeTrue())
		})

		It("matches new object", func() {
			updateEvent.MetaOld = badMeta
			updateEvent.MetaNew = matchingMeta
			Expect(pred.Update(updateEvent)).To(BeTrue())
		})

		It("matches both objects", func() {
			updateEvent.MetaOld = matchingMeta
			updateEvent.MetaNew = matchingMeta
			Expect(pred.Update(updateEvent)).To(BeTrue())
		})

		It("matches neither object", func() {
			updateEvent.MetaOld = badMeta
			updateEvent.MetaNew = badMeta
			Expect(pred.Update(updateEvent)).To(BeFalse())
		})
	})
})
