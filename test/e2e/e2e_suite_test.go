package e2e_test

import (
	"context"
	"strings"
	"testing"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/supergloo/pkg/setup"
	"github.com/solo-io/supergloo/test/testutils"
	v1 "k8s.io/api/core/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var (
	kube    kubernetes.Interface
	rootCtx context.Context
	cancel  func()
)

var _ = BeforeSuite(func() {
	rootCtx, cancel = context.WithCancel(context.TODO())
	kube = testutils.MustKubeClient()
	// create sg ns
	_, err := kube.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "supergloo-system"},
	})
	Expect(err).NotTo(HaveOccurred())

	// start supergloo
	go func() {
		defer GinkgoRecover()
		err := setup.Main(rootCtx, func(e error) {
			Expect(e).NotTo(HaveOccurred())
		})
		Expect(err).NotTo(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {
	cancel()
	kube.CoreV1().Namespaces().Delete("supergloo-system", nil)
	kube.CoreV1().Namespaces().Delete("istio-system", nil)
	clusterroles, err := kube.RbacV1beta1().ClusterRoles().List(metav1.ListOptions{})
	if err == nil {
		for _, cr := range clusterroles.Items {
			if strings.Contains(cr.Name, "istio") {
				kube.RbacV1beta1().ClusterRoles().Delete(cr.Name, nil)
			}
		}
	}
	clusterrolebindings, err := kube.RbacV1beta1().ClusterRoleBindings().List(metav1.ListOptions{})
	if err == nil {
		for _, cr := range clusterrolebindings.Items {
			if strings.Contains(cr.Name, "istio") {
				kube.RbacV1beta1().ClusterRoleBindings().Delete(cr.Name, nil)
			}
		}
	}
	webhooks, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().List(metav1.ListOptions{})
	if err == nil {
		for _, wh := range webhooks.Items {
			if strings.Contains(wh.Name, "istio") {
				kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(wh.Name, nil)
			}
		}
	}

	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())

	exts, err := apiexts.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	crds, err := exts.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err == nil {
		for _, cr := range crds.Items {
			if strings.Contains(cr.Name, "istio") {
				exts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(cr.Name, nil)
			}
		}
	}
})
