package utils

import (
	"github.com/solo-io/supergloo/pkg/install/shared"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func DeployFromYamlWithIstioInject(cfg *rest.Config, namespace, yamlManifest string) error {
	injected, err := IstioInject(namespace, yamlManifest)
	if err != nil {
		return err
	}
	return DeployFromYaml(cfg, namespace, injected)
}

func DeployFromYaml(cfg *rest.Config, namespace, yamlManifest string) error {
	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	apiext, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	installer := shared.NewKubeInstaller(kube, apiext, namespace)

	kubeObjs, err := shared.ParseKubeManifest(yamlManifest)
	if err != nil {
		return err
	}

	for _, kubeOjb := range kubeObjs {
		if err := installer.Create(kubeOjb); err != nil {
			return err
		}
	}
	return nil
}
