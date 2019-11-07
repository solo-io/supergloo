package main

import (
	"fmt"

	"github.com/solo-io/go-utils/kubeutils"
	mcv1 "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"sigs.k8s.io/yaml"
)

func main() {
	kubeConfig, err := kubeutils.GetKubeConfig("", "")
	if err != nil {
		panic(err)
	}
	kubeCfgSecret1, err := secretconverter.KubeConfigToSecret(&v1.KubeConfig{
		KubeConfig: mcv1.KubeConfig{
			Metadata: core.Metadata{Name: "kubeconfig1"},
			Config:   *kubeConfig,
			Cluster:  "linkerd-mesh-bridge",
		},
	})
	byt, err := yaml.Marshal(kubeCfgSecret1)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(byt))
}
