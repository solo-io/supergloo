package testutils

import (
	"github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

func MustKubeClient() kubernetes.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return kubeClient
}


func MustApiExtsClient() apiexts.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	apiExtsClient, err := apiexts.NewForConfig(restConfig)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return apiExtsClient
}
