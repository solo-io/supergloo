package testutils

import (
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var istioInstalledCrds = []string{
	"meshpolicies.authentication.istio.io",
	"policies.authentication.istio.io",
}

func installedByIstio(crdName string) bool {
	for _, n := range istioInstalledCrds {
		if crdName == n {
			return true
		}
	}
	return false
}

func WaitForIstioTeardown(ns string) {
	EventuallyWithOffset(1, func() []kubev1.Service {
		svcs, err := MustKubeClient().CoreV1().Services(ns).List(v1.ListOptions{})
		if err != nil {
			// namespace is gone
			return []kubev1.Service{}
		}
		return svcs.Items
	}, time.Second*30).Should(BeEmpty())
}

func TeardownIstio(kube kubernetes.Interface) {
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
}
