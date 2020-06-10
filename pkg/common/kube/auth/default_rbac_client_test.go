package auth_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/auth"
	v1 "k8s.io/api/core/v1"
	rbacapiv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("default rbac client", func() {
	var (
		rbacClient    auth.RbacClient
		fakeClientset *fake.Clientset

		testErr = eris.New("hello")
		name    = "hello"
	)

	BeforeEach(func() {
		fakeClientset = fake.NewSimpleClientset()
		rbacClient = auth.RbacClientProvider(fakeClientset)
	})

	It("will attempt to update if create fails with already exist", func() {
		sa := &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: name,
			},
		}
		roles := []*rbacapiv1.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			},
		}
		fakeClientset.PrependReactor("update", "clusterrolebindings",
			func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, testErr
			})
		_, err := fakeClientset.RbacV1().ClusterRoleBindings().Create(&rbacapiv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s-clusterrole-binding", sa.GetName(), name)},
		})
		Expect(err).NotTo(HaveOccurred())
		err = rbacClient.BindClusterRolesToServiceAccount(sa, roles)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(testErr))

	})

	It("will return result if create fails with any other error", func() {
		sa := &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: name,
			},
		}
		roles := []*rbacapiv1.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			},
		}

		err := rbacClient.BindClusterRolesToServiceAccount(sa, roles)
		Expect(err).NotTo(HaveOccurred())
		_, err = fakeClientset.RbacV1().ClusterRoleBindings().Get(
			fmt.Sprintf("%s-%s-clusterrole-binding", sa.GetName(), name),
			metav1.GetOptions{},
		)
		Expect(err).NotTo(HaveOccurred())
	})
})
