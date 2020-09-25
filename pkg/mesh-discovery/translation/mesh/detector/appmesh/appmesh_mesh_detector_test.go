package appmesh_test

import (
	"github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	. "github.com/onsi/ginkgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AppMesh MeshDetector", func() {

	meshNamespace := "namespace-one"
	meshName := "mesh-name"

	var _ = v1beta2.Mesh{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: meshNamespace,
			Name:      meshName,
		},
		Spec: v1beta2.MeshSpec{
			AWSName: &meshName,
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels:      map[string]string{"mesh": meshName},
				MatchExpressions: nil,
			},
			EgressFilter: nil,
			MeshOwner:    nil,
		},
		Status: v1beta2.MeshStatus{},
	}

})
