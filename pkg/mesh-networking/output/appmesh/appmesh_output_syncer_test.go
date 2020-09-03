package appmesh_test

import (
	"context"
	"log"

	"github.com/aws/aws-app-mesh-controller-for-k8s/pkg/aws/services"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	appmeshsdk "github.com/aws/aws-sdk-go/service/appmesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/external-apis/pkg/api/appmesh/appmesh.aws.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/external-apis/pkg/api/appmesh/appmesh.aws.solo.io/v1alpha1/sets"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/external/appmesh"
	"github.com/solo-io/skv2/contrib/pkg/output/errhandlers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/output/appmesh"
)

var _ = Describe("AppmeshOutputSyncer", func() {
	var (
		appMesh  services.AppMesh
		meshName = "test-mesh-" + testutils.RandString(4)
	)
	BeforeEach(func() {
		sess, err := session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			Skip("failed to load AWS credentials: " + err.Error() + ". AWS integration test will be skipped:")
		}

		appMesh = services.NewAppMesh(sess)

		_, err = appMesh.CreateMesh(&appmeshsdk.CreateMeshInput{
			MeshName: aws.String(meshName),
			Spec:     &appmeshsdk.MeshSpec{},
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		if appMesh == nil {
			return
		}
		_, err := appMesh.DeleteMesh(&appmeshsdk.DeleteMeshInput{MeshName: aws.String(meshName)})
		if err != nil {
			log.Printf("WARNING: failed to clean up AppMesh test Mesh %v: %v", meshName, err)
		}
	})
	It("syncs resources with AWS AppMesh", func() {
		client := appmesh.NewClient("", appMesh)
		syncer := NewOutputSyncer(client)

		errHandler := errhandlers.AppendingErrHandler{}

		virtualNodeSet := v1alpha1sets.NewVirtualNodeSet(&v1alpha1.VirtualNode{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-virtual-node",
				Namespace: meshName,
			},
			Spec: appmeshsdk.VirtualNodeSpec{
				BackendDefaults:  nil,
				Backends:         nil,
				Listeners:        nil,
				Logging:          nil,
				ServiceDiscovery: nil,
			},
		})

		outputs := appmesh.NewSnapshot(
			virtualNodeSet,
		)
		err := syncer.Apply(context.TODO(), outputs, errHandler)
		Expect(err).NotTo(HaveOccurred())

	})
})
