package utils

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/log"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)
/*


 */
func DeployPrometheus(namespace, deploymentName, configmapName string, port uint32, kube kubernetes.Interface) error {
	deploymentYaml, err := BasicPrometheusDeployment(namespace, deploymentName, configmapName)
	if err != nil {
		return errors.Wrapf(err, "internal error") // should never happen
	}
	log.Printf(deploymentYaml)
	var prometheusDeployment v1beta1.Deployment
	if err := yaml.Unmarshal([]byte(deploymentYaml), &prometheusDeployment); err != nil {
		return errors.Wrapf(err, "internal error") // should never happen
	}
	prometheusDeployment.Namespace = namespace
	prometheusDeployment.Name = deploymentName
	if _, err := kube.ExtensionsV1beta1().Deployments(namespace).Create(&prometheusDeployment); err != nil {
		return err
	}
	serviceYaml, err := BasicPrometheusService(namespace, deploymentName, port)
	if err != nil {
		return errors.Wrapf(err, "internal error") // should never happen
	}
	var prometheusService v1.Service
	if err := yaml.Unmarshal([]byte(serviceYaml), &prometheusService); err != nil {
		return errors.Wrapf(err, "internal error") // should never happen
	}
	prometheusService.Namespace = namespace
	prometheusService.Name = deploymentName
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
			// Required by Solo-Kit; TODO: consider removing this requirement
			Annotations: map[string]string{"resource_kind": "*v1.Config"},
		},
		Data: map[string]string{"prometheus.yml": BasicPrometheusConfig},
	})
	return err
}
