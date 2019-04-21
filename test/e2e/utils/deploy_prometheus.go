package utils

import (
	"time"

	"k8s.io/client-go/kubernetes"

	sgtestutils "github.com/solo-io/supergloo/test/testutils"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DeployPrometheus(kube kubernetes.Interface, namespace string) error {
	_, err := kube.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	})
	if err != nil {
		return err
	}

	manifest, err := HelmTemplate("--name=prometheus",
		"--namespace="+namespace,
		"--set", "rbac.create=true",
		"--set", "server.persistentVolume.enabled=false",
		"--set", "alertmanager.enabled=false",
		MustTestFile("prometheus-8.9.0.tgz"))
	if err != nil {
		return err
	}

	err = KubectlApply(namespace, manifest)
	if err != nil {
		return err
	}

	Eventually(func() error {
		_, err := kube.ExtensionsV1beta1().Deployments(namespace).Get("prometheus-server", metav1.GetOptions{})
		return err
	}, time.Minute*2).ShouldNot(HaveOccurred())

	return sgtestutils.WaitUntilPodsRunning(time.Minute, namespace, "prometheus-server")
}

func TeardownPrometheus(kube kubernetes.Interface, namespace string) error {
	manifest, err := HelmTemplate("--name=prometheus",
		"--namespace="+namespace,
		"--set", "rbac.create=true",
		"--set", "server.persistentVolume.enabled=false",
		"--set", "alertmanager.enabled=false",
		MustTestFile("prometheus-8.9.0.tgz"))
	if err != nil {
		return err
	}

	err = KubectlDelete(namespace, manifest)
	if err != nil {
		return err
	}

	err = kube.CoreV1().Namespaces().Delete(namespace, nil)
	if err != nil {
		return err
	}

	return nil
}
