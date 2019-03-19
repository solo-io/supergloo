package istio_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/registration/istio"
	"github.com/solo-io/supergloo/test/inputs"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("IstioSecretDeleter", func() {
	var kube kubernetes.Interface
	var ns string
	BeforeEach(func() {
		kube = fake.NewSimpleClientset()
		ns = "istio-was-installed-here"
		kube.CoreV1().Namespaces().Create(&kubev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	})
	Context("when it receives a mesh with a custom root cert", func() {
		BeforeEach(func() {
			for _, secret := range []*kubev1.Secret{
				{ObjectMeta: metav1.ObjectMeta{Name: "istio.a", Namespace: ns}},
				{ObjectMeta: metav1.ObjectMeta{Name: "istio.b", Namespace: ns}},
				{ObjectMeta: metav1.ObjectMeta{Name: "istio.default", Namespace: ns}},
				{ObjectMeta: metav1.ObjectMeta{Name: "notistio", Namespace: ns}},
			} {
				_, err := kube.CoreV1().Secrets(ns).Create(secret)
				Expect(err).NotTo(HaveOccurred())
			}
		})
		It("deletes all secrets in the istio installnamespace with the prefix istio.", func() {
			err := NewIstioSecretDeleter(kube).Sync(context.TODO(), &v1.RegistrationSnapshot{
				Meshes: v1.MeshesByNamespace{"": v1.MeshList{inputs.IstioMeshWithInstallNs(ns, ns, &core.ResourceRef{"custom", "rootcert"})}},
			})
			Expect(err).NotTo(HaveOccurred())
			secrets, err := kube.CoreV1().Secrets(ns).List(metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(secrets.Items).To(Equal([]kubev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "notistio",
						Namespace: "istio-was-installed-here",
					},
				},
			}))
		})
	})
})
