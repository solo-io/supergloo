package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)



func DeployPrometheus(namespace, name string, kube kubernetes.Interface) error {
	_, err := kube.CoreV1().ConfigMaps(namespace).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{"prometheus.yaml": demoIstioPrometheusDeployment},
	})
	return err
}

const demoIstioPrometheusDeployment =""


func DeployPrometheusConfigmap(namespace, name string, kube kubernetes.Interface) error {
	_, err := kube.CoreV1().ConfigMaps(namespace).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{"prometheus.yaml": demoIstioPrometheusConfig},
	})
	return err
}

