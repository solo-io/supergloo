package setup

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitCache(opts *options.Options) error {

	kube, err := common.GetKubernetesClient()
	if err != nil {
		return err
	}
	opts.Cache.KubeClient = kube

	list, err := kube.CoreV1().Namespaces().List(v1.ListOptions{IncludeUninitialized: false})
	if err != nil {
		return err
	}
	var namespaces = []string{}
	for _, ns := range list.Items {
		namespaces = append(namespaces, ns.ObjectMeta.Name)
	}
	opts.Cache.Namespaces = namespaces

	return nil
}
