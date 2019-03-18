package testutils

import (
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WaitForNamespaceTeardown(ns string) {
	EventuallyWithOffset(1, func() (bool, error) {
		namespaces, err := MustKubeClient().CoreV1().Namespaces().List(v1.ListOptions{})
		if err != nil {
			// namespace is gone
			return false, err
		}
		for _, n := range namespaces.Items {
			if n.Name == ns {
				return false, nil
			}
		}
		return true, nil
	}, time.Second*30).Should(BeTrue())
}

func TeardownSuperGloo(kube kubernetes.Interface) {
	kube.CoreV1().Namespaces().Delete("supergloo-system", nil)
	clusterroles, err := kube.RbacV1beta1().ClusterRoles().List(metav1.ListOptions{})
	if err == nil {
		for _, cr := range clusterroles.Items {
			if strings.Contains(cr.Name, "supergloo") {
				kube.RbacV1beta1().ClusterRoles().Delete(cr.Name, nil)
			}
		}
	}
	clusterrolebindings, err := kube.RbacV1beta1().ClusterRoleBindings().List(metav1.ListOptions{})
	if err == nil {
		for _, cr := range clusterrolebindings.Items {
			if strings.Contains(cr.Name, "supergloo") {
				kube.RbacV1beta1().ClusterRoleBindings().Delete(cr.Name, nil)
			}
		}
	}
	webhooks, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().List(metav1.ListOptions{})
	if err == nil {
		for _, wh := range webhooks.Items {
			if strings.Contains(wh.Name, "supergloo") {
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
			if strings.Contains(cr.Name, "supergloo") {
				exts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(cr.Name, nil)
			}
		}
	}
}
