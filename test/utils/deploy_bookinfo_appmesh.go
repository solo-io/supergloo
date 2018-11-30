package utils

import (
	"github.com/solo-io/supergloo/pkg/install/shared"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func DeployBookinfoAppMesh(cfg *rest.Config, namespace, meshName, virtualNodeName, awsRegion string) error {
	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	apiext, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	installer := shared.NewKubeInstaller(kube, apiext, namespace)

	kubeObjs, err := shared.ParseKubeManifest(IstioBookinfoYaml)
	if err != nil {
		return err
	}

	AwsAppMeshInjectKubeObjList(kubeObjs, meshName, virtualNodeName, awsRegion, []uint32{9080})

	for _, kubeOjb := range kubeObjs {
		if err := installer.Create(kubeOjb); err != nil {
			return err
		}
	}
	return nil
}
