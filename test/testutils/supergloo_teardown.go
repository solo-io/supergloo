package testutils

import (
	"log"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	kubev1 "k8s.io/api/core/v1"
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
	}, time.Second*180).Should(BeTrue())
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

// remove supergloo controller pod(s)
func DeleteSuperglooPods(kube kubernetes.Interface, superglooNamespace string) {
	podsToDelete := []string{
		"supergloo",
		"mesh-discovery",
	}
	for _, pod := range podsToDelete {
		// wait until pod is gone
		Eventually(func() error {
			dep, err := kube.ExtensionsV1beta1().Deployments(superglooNamespace).Get(pod, metav1.GetOptions{})
			if err != nil {
				return err
			}
			dep.Spec.Replicas = proto.Int(0)
			_, err = kube.ExtensionsV1beta1().Deployments(superglooNamespace).Update(dep)
			if err != nil {
				return err
			}
			pods, err := kube.CoreV1().Pods(superglooNamespace).List(metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, p := range pods.Items {
				if strings.HasPrefix(p.Name, pod) {
					return errors.Errorf("%s pods still exist", pod)
				}
			}
			return nil
		}, time.Second*120).ShouldNot(HaveOccurred())
	}

}

func WaitUntilPodsRunning(timeout time.Duration, namespace string, podPrefixes ...string) error {
	pods := clients.MustKubeClient().CoreV1().Pods(namespace)
	podsWithPrefixReady := func(prefix string) (bool, error) {
		list, err := pods.List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		var podsWithPrefix []kubev1.Pod
		for _, pod := range list.Items {
			if strings.HasPrefix(pod.Name, prefix) {
				podsWithPrefix = append(podsWithPrefix, pod)
			}
		}
		if len(podsWithPrefix) == 0 {
			return false, errors.Errorf("no pods found with prefix %v", prefix)
		}
		for _, pod := range podsWithPrefix {
			var podReady bool
			for _, cond := range pod.Status.Conditions {
				if cond.Type == kubev1.ContainersReady && cond.Status == kubev1.ConditionTrue {
					podReady = true
					break
				}
			}
			if !podReady {
				return false, nil
			}
		}
		return true, nil
	}
	failed := time.After(timeout)
	notYetRunning := make(map[string]struct{})
	for {
		select {
		case <-failed:
			return errors.Errorf("timed out waiting for pods to come online: %v", notYetRunning)
		case <-time.After(time.Second / 2):
			notYetRunning = make(map[string]struct{})
			for _, prefix := range podPrefixes {
				ready, err := podsWithPrefixReady(prefix)
				if err != nil {
					log.Printf("failed to get pod status: %v", err)
					notYetRunning[prefix] = struct{}{}
				}
				if !ready {
					notYetRunning[prefix] = struct{}{}
				}
			}
			if len(notYetRunning) == 0 {
				return nil
			}
		}

	}
}
