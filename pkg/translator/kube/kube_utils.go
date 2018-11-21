package kube

import (
	"github.com/solo-io/solo-kit/pkg/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// Note: This assumes the pod will get restarted automatically due to the kubernetes deployment spec
func RestartPods(kube kubernetes.Interface, namespace string, selector map[string]string) error {
	if kube == nil {
		return errors.Errorf("kubernetes suppport is currently disabled. see SuperGloo documentation" +
			" for utilizing pod restarts")
	}
	if err := kube.CoreV1().Pods(namespace).DeleteCollection(nil, kubemeta.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector).String(),
	}); err != nil {
		return errors.Wrapf(err, "restarting pods with selector %v", selector)
	}
	return nil
}
