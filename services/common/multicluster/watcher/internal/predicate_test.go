package internal_watcher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	. "github.com/solo-io/mesh-projects/services/common/multicluster/watcher/internal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("common", func() {
	Context("hasRequiredMetadata", func() {
		var meta metav1.Object
		It("will not work with just labels", func() {
			meta = &metav1.ObjectMeta{
				Labels: map[string]string{multicluster.MultiClusterLabel: "true"},
			}
			Expect(HasRequiredMetadata(meta)).To(BeFalse())
		})
		It("will not work with just namespace", func() {
			meta = &metav1.ObjectMeta{
				Namespace: env.DefaultWriteNamespace,
			}
			Expect(HasRequiredMetadata(meta)).To(BeFalse())
		})
		It("will not work with one condition", func() {
			meta = &metav1.ObjectMeta{
				Labels:    map[string]string{multicluster.MultiClusterLabel: "true"},
				Namespace: env.DefaultWriteNamespace,
			}
			Expect(HasRequiredMetadata(meta)).To(BeTrue())
		})
	})
	Context("will fire an update event if either new or old matches", func() {
		var (
			updateEvent event.UpdateEvent
			pred        *MultiClusterPredicate

			matchingMeta = &metav1.ObjectMeta{
				Labels:    map[string]string{multicluster.MultiClusterLabel: "true"},
				Namespace: env.DefaultWriteNamespace,
			}

			badMeta = &metav1.ObjectMeta{
				Labels: map[string]string{multicluster.MultiClusterLabel: "true"},
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
