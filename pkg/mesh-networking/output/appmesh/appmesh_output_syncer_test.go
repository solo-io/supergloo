package appmesh_test

import (
	"github.com/aws/aws-app-mesh-controller-for-k8s/pkg/aws/services"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/output/appmesh"
)

var _ = Describe("AppmeshOutputSyncer", func() {
	var (
		appMesh services.AppMesh
	)
	BeforeEach(func() {


	})
	It("syncs resources with AWS AppMesh", func() {

	})
})
