package kube

import (
	"github.com/openshift/api/security/v1"
	security "github.com/openshift/client-go/security/clientset/versioned"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// If you change this interface, you have to rerun mockgen
type SecurityClient interface {
	GetScc(name string) (*v1.SecurityContextConstraints, error)
	UpdateScc(*v1.SecurityContextConstraints) (*v1.SecurityContextConstraints, error)
}

type KubeSecurityClient struct {
	securityClient *security.Clientset
}

func (client *KubeSecurityClient) GetScc(name string) (*v1.SecurityContextConstraints, error) {
	return client.securityClient.SecurityV1().SecurityContextConstraints().Get(name, kubemeta.GetOptions{})
}

func (client *KubeSecurityClient) UpdateScc(scc *v1.SecurityContextConstraints) (*v1.SecurityContextConstraints, error) {
	return client.securityClient.SecurityV1().SecurityContextConstraints().Update(scc)
}
