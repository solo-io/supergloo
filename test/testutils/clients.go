package testutils

import (
	"github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/client-go/kubernetes"
)

func MustKubeClient() kubernetes.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return kubeClient
}
