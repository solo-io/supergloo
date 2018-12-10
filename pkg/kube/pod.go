package kube

import (
	"github.com/solo-io/solo-kit/pkg/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type PodClient interface {
	RestartPods(namespace string, selector map[string]string) error
}

type KubePodClient struct {
	kube kubernetes.Interface
}

func NewKubePodClient(kube kubernetes.Interface) *KubePodClient {
	return &KubePodClient{
		kube: kube,
	}
}

// Note: This assumes the pod will get restarted automatically due to the kubernetes deployment spec
func (client *KubePodClient) RestartPods(namespace string, selector map[string]string) error {
	if client.kube == nil {
		return errors.Errorf("kubernetes suppport is currently disabled. see SuperGloo documentation" +
			" for utilizing pod restarts")
	}
	if err := client.kube.CoreV1().Pods(namespace).DeleteCollection(nil, kubemeta.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector).String(),
	}); err != nil {
		return errors.Wrapf(err, "restarting pods with selector %v", selector)
	}
	return nil
}
