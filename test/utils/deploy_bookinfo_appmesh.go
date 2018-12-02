package utils

import (
	"github.com/solo-io/solo-kit/pkg/utils/nameutils"
	"github.com/solo-io/supergloo/pkg/install/shared"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

func DeployBookinfoAppMesh(cfg *rest.Config, namespace, meshName, awsRegion string) error {
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

	for i, obj := range kubeObjs {
		if deployment, ok := obj.(*v1beta1.Deployment); ok {
			nameDeployment := strings.TrimSuffix(deployment.Name, "-v1")
			nameDeployment = strings.TrimSuffix(nameDeployment, "-v2")
			nameDeployment = strings.TrimSuffix(nameDeployment, "-v3")
			host := nameDeployment + "." + namespace + ".svc.cluster.local"
			virtualNodeName := nameutils.SanitizeName(host)
			AwsAppMeshInjectKubeObjList(shared.KubeObjectList{deployment}, meshName, virtualNodeName, awsRegion, []uint32{9080})
			kubeObjs[i] = deployment
		}
	}

	for _, kubeOjb := range kubeObjs {
		if err := installer.Create(kubeOjb); err != nil {
			return err
		}
	}
	return nil
}
