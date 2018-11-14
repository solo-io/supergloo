package utils

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeployPrometheus(namespace string, kube kubernetes.Interface) error {
	var prometheusDeployment v1beta1.Deployment
	if err := yaml.Unmarshal([]byte(BasicPrometheusDeployment), &prometheusDeployment); err != nil {
		return errors.Wrapf(err, "internal error") // should never happen
	}
	prometheusDeployment.Namespace = namespace
	if _, err := kube.ExtensionsV1beta1().Deployments(namespace).Create(&prometheusDeployment); err != nil {
		return err
	}
	var prometheusService v1.Service
	err := yaml.Unmarshal([]byte(IstioPrometheusService), &prometheusService)
	if err != nil {
		return errors.Wrapf(err, "internal error") // should never happen
	}
	prometheusService.Namespace = namespace
	if _, err := kube.CoreV1().Services(namespace).Create(&prometheusService); err != nil {
		return err
	}
	return nil
}

func DeployPrometheusConfigmap(namespace, name string, kube kubernetes.Interface) error {
	_, err := kube.CoreV1().ConfigMaps(namespace).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{"prometheus.yaml": BasicPrometheusConfig},
	})
	return err
}
