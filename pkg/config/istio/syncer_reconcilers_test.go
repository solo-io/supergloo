package istio_test

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"

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
		tlsSecretClient          v1.TlsSecretClient

		rec istio.Reconcilers
	)
	BeforeEach(func() {
		mem := &factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()}

		rbacConfigClientLoader := func() (rbacv1alpha1.RbacConfigClient, error) {
			return rbacv1alpha1.NewRbacConfigClient(mem)
		}

		serviceRoleClientLoader := func() (rbacv1alpha1.ServiceRoleClient, error) {
			return rbacv1alpha1.NewServiceRoleClient(mem)
		}

		serviceRoleBindingClientLoader := func() (rbacv1alpha1.ServiceRoleBindingClient, error) {
			return rbacv1alpha1.NewServiceRoleBindingClient(mem)
		}

		meshPolicyClientLoader := func() (policyv1alpha1.MeshPolicyClient, error) {
			return policyv1alpha1.NewMeshPolicyClient(mem)
		}

		destinationRuleClientLoader := func() (v1alpha3.DestinationRuleClient, error) {
			return v1alpha3.NewDestinationRuleClient(mem)
		}

		virtualServiceClientLoader := func() (v1alpha3.VirtualServiceClient, error) {
			return v1alpha3.NewVirtualServiceClient(mem)
		}

		var err error
		tlsSecretClient, err = v1.NewTlsSecretClient(mem)
		Expect(err).NotTo(HaveOccurred())
		tlsSecretClientReconciler := v1.NewTlsSecretReconciler(tlsSecretClient)

		labels := map[string]string{"test": "labels"}
		rec = istio.NewIstioReconcilers(labels,
			rbacConfigClientLoader,
			serviceRoleClientLoader,
			serviceRoleBindingClientLoader,
			meshPolicyClientLoader,
			destinationRuleClientLoader,
			virtualServiceClientLoader,
			tlsSecretClientReconciler,
		)

		rbacConfigClient, _ = rbacConfigClientLoader()
		serviceRoleClient, _ = serviceRoleClientLoader()
		serviceRoleBindingClient, _ = serviceRoleBindingClientLoader()
		meshPolicyClient, _ = meshPolicyClientLoader()
		destinationRuleClient, _ = destinationRuleClientLoader()
		virtualServiceClient, _ = virtualServiceClientLoader()
	})

	It("reconciles from a set of MeshConfig", func() {
		writeNs := "write-namespace"
		mpMeta, drMeta, vsMeta, rbMeta, srMeta, srbMeta, tlsMeta := randomMeta(""), randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs), randomMeta(writeNs)
		config := &MeshConfig{
			MeshPolicy:       &policyv1alpha1.MeshPolicy{Metadata: mpMeta},
			DestinationRules: []*v1alpha3.DestinationRule{{Metadata: drMeta}},
			VirtualServices:  []*v1alpha3.VirtualService{{Metadata: vsMeta}},
			SecurityConfig: SecurityConfig{
				ServiceRoles:        []*rbacv1alpha1.ServiceRole{{Metadata: srMeta}},
				ServiceRoleBindings: []*rbacv1alpha1.ServiceRoleBinding{{Metadata: srbMeta}},
				RbacConfig:          &rbacv1alpha1.RbacConfig{Metadata: rbMeta},
			},
			RootCert: &v1.TlsSecret{Metadata: tlsMeta},
		}

		err := rec.ReconcileAll(context.TODO(), config)
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

		_, err = tlsSecretClient.Read(tlsMeta.Namespace, tlsMeta.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
	})
})

func randomMeta(writeNamespace string) core.Metadata {
	return core.Metadata{Name: "a" + helpers.RandString(4), Namespace: writeNamespace}
}
