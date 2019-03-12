package istio_test

import (
	"context"

	"github.com/solo-io/supergloo/pkg/config/istio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	. "github.com/solo-io/supergloo/pkg/translator/istio"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	policyv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
)

var _ = Describe("SyncerReconcilers", func() {
	var (
		rbacConfigClient         rbacv1alpha1.RbacConfigClient
		serviceRoleClient        rbacv1alpha1.ServiceRoleClient
		serviceRoleBindingClient rbacv1alpha1.ServiceRoleBindingClient
		meshPolicyClient         policyv1alpha1.MeshPolicyClient
		destinationRuleClient    v1alpha3.DestinationRuleClient
		virtualServiceClient     v1alpha3.VirtualServiceClient

		rec istio.Reconcilers
	)
	BeforeEach(func() {
		mem := &factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()}

		var err error
		rbacConfigClient, err = rbacv1alpha1.NewRbacConfigClient(mem)
		Expect(err).NotTo(HaveOccurred())
		rbacConfigReconciler := rbacv1alpha1.NewRbacConfigReconciler(rbacConfigClient)

		serviceRoleClient, err = rbacv1alpha1.NewServiceRoleClient(mem)
		Expect(err).NotTo(HaveOccurred())
		serviceRoleReconciler := rbacv1alpha1.NewServiceRoleReconciler(serviceRoleClient)

		serviceRoleBindingClient, err = rbacv1alpha1.NewServiceRoleBindingClient(mem)
		Expect(err).NotTo(HaveOccurred())
		serviceRoleBindingReconciler := rbacv1alpha1.NewServiceRoleBindingReconciler(serviceRoleBindingClient)

		meshPolicyClient, err = policyv1alpha1.NewMeshPolicyClient(mem)
		Expect(err).NotTo(HaveOccurred())
		meshPolicyReconciler := policyv1alpha1.NewMeshPolicyReconciler(meshPolicyClient)

		destinationRuleClient, err = v1alpha3.NewDestinationRuleClient(mem)
		Expect(err).NotTo(HaveOccurred())
		destinationRuleReconciler := v1alpha3.NewDestinationRuleReconciler(destinationRuleClient)

		virtualServiceClient, err = v1alpha3.NewVirtualServiceClient(mem)
		Expect(err).NotTo(HaveOccurred())
		virtualServiceReconciler := v1alpha3.NewVirtualServiceReconciler(virtualServiceClient)

		labels := map[string]string{"test": "labels"}
		rec = istio.NewIstioReconcilers(labels,
			rbacConfigReconciler,
			serviceRoleReconciler,
			serviceRoleBindingReconciler,
			meshPolicyReconciler,
			destinationRuleReconciler,
			virtualServiceReconciler,
		)
	})

	It("reconciles from a set of MeshConfig", func() {
		writeNs := "write-namespace"
		mpMeta, drMeta, vsMeta, rbMeta, srMeta, srbMeta := randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs)
		config := &MeshConfig{
			MeshPolicy:       &policyv1alpha1.MeshPolicy{Metadata: mpMeta},
			DestinationRules: []*v1alpha3.DestinationRule{{Metadata: drMeta}},
			VirtualServices:  []*v1alpha3.VirtualService{{Metadata: vsMeta}},
			SecurityConfig: SecurityConfig{
				ServiceRoles:        []*rbacv1alpha1.ServiceRole{{Metadata: srMeta}},
				ServiceRoleBindings: []*rbacv1alpha1.ServiceRoleBinding{{Metadata: srbMeta}},
				RbacConfig:          &rbacv1alpha1.RbacConfig{Metadata: rbMeta},
			},
		}

		err := rec.ReconcileAll(context.TODO(), writeNs, config)
		Expect(err).NotTo(HaveOccurred())

		// reconciler should have created each type of resource
		_, err = meshPolicyClient.Read(mpMeta.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		_, err = rbacConfigClient.Read(rbMeta.Namespace, rbMeta.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		_, err = destinationRuleClient.Read(drMeta.Namespace, drMeta.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		_, err = virtualServiceClient.Read(vsMeta.Namespace, vsMeta.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		_, err = serviceRoleClient.Read(srMeta.Namespace, srMeta.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		_, err = serviceRoleBindingClient.Read(srbMeta.Namespace, srbMeta.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
	})
})

func randomMeta(writeNamespace string) core.Metadata {
	return core.Metadata{Name: "a" + helpers.RandString(4), Namespace: writeNamespace}
}
