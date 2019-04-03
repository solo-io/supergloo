package utils

import (
	"github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateNs(ns string) error {
	kube := clients.MustKubeClient()
	_, err := kube.CoreV1().Namespaces().Create(&kubev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	})

	return err
}

func MustCreateNs(ns string) {
	gomega.ExpectWithOffset(1, CreateNs(ns)).NotTo(gomega.HaveOccurred())
}

func DeleteNs(ns string) error {
	kube := clients.MustKubeClient()
	err := kube.CoreV1().Namespaces().Delete(ns, nil)

	return err
}

func MustDeleteNs(ns string) {
	gomega.ExpectWithOffset(1, DeleteNs(ns)).NotTo(gomega.HaveOccurred())
}

func ConfigMap(ns, name, data string, labels map[string]string) kubev1.ConfigMap {
	return kubev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Data: map[string]string{"data": data},
	}
}

func CreateConfigMap(cm kubev1.ConfigMap) error {
	kube := clients.MustKubeClient()
	_, err := kube.CoreV1().ConfigMaps(cm.Namespace).Create(&cm)

	return err
}

func MustCreateConfigMap(cm kubev1.ConfigMap) {
	gomega.ExpectWithOffset(1, CreateConfigMap(cm)).NotTo(gomega.HaveOccurred())
}
