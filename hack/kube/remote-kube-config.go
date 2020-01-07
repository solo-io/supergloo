package main

import (
	"flag"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/kubeutils"
	mcv1 "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
)

func main() {
	namespace := flag.String("n", "", "secret name")
	secretName := flag.String("s", "remote-kube-config", "namespace")
	filepath := flag.String("f", "", "kube config filepath")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		panic("cluster name must be passed as first argument")
	}
	clusterName := args[0]
	kubeConfig, err := kubeutils.GetKubeConfig("", *filepath)
	if err != nil {
		panic(err)
	}
	kubeCfgSecret1, err := secretconverter.KubeConfigToSecret(&v1.KubeConfig{
		KubeConfig: mcv1.KubeConfig{
			Metadata: core.Metadata{
				Name:      *secretName,
				Namespace: *namespace,
			},
			Config:  *kubeConfig,
			Cluster: clusterName,
		},
	})
	kubeCfgSecret1.Kind = "Secret"
	kubeCfgSecret1.APIVersion = "v1"
	byt, err := yaml.Marshal(kubeCfgSecret1)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(byt))
}
