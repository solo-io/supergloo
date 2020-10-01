package appmesh_test

import (
	"context"
	"fmt"

	aws_v1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	input "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	detector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/appmesh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AppMesh MeshDetector", func() {

	var (
		clusterName = "cluster1"
		meshName    = "mesh-name"
		meshArn     = "arn:aws:appmesh:us-east-2:1234:mesh/mesh-name"

		meshDetector detector.MeshDetector
	)

	BeforeEach(func() {
		meshDetector = appmesh.NewMeshDetector(context.Background())
	})

	It("detects an App Mesh mesh from a well-formed mesh resource", func() {
		awsMesh := aws_v1beta2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:        meshName,
				ClusterName: clusterName,
			},
			Spec: aws_v1beta2.MeshSpec{
				AWSName: &meshName,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"mesh": meshName},
				},
			},
			Status: aws_v1beta2.MeshStatus{
				MeshARN: &meshArn,
			},
		}

		expected := v1alpha2.MeshSlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s", meshName, clusterName),
					Namespace: "service-mesh-hub",
					Labels: map[string]string{
						"owner.discovery.smh.solo.io":   "service-mesh-hub",
						"cluster.discovery.smh.solo.io": "cluster1",
					},
				},
				Spec: v1alpha2.MeshSpec{
					MeshType: &v1alpha2.MeshSpec_AwsAppMesh_{
						AwsAppMesh: &v1alpha2.MeshSpec_AwsAppMesh{
							AwsName:      meshName,
							Region:       "us-east-2",
							AwsAccountId: "1234",
							Arn:          meshArn,
							// TODO investigate multicluster app mesh
							Clusters: []string{clusterName},
						},
					},
				},
			},
		}

		builder := input.NewInputSnapshotManualBuilder("app mesh test")
		builder.AddMeshes([]*aws_v1beta2.Mesh{&awsMesh})

		actual, err := meshDetector.DetectMeshes(builder.Build())
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})

	It("does not detect meshes that haven't been assigned a name", func() {
		awsMesh := aws_v1beta2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:        meshName,
				ClusterName: clusterName,
			},
			Spec: aws_v1beta2.MeshSpec{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"mesh": meshName},
				},
			},
			Status: aws_v1beta2.MeshStatus{
				MeshARN: &meshArn,
			},
		}

		builder := input.NewInputSnapshotManualBuilder("app mesh test")
		builder.AddMeshes([]*aws_v1beta2.Mesh{&awsMesh})

		actual, err := meshDetector.DetectMeshes(builder.Build())
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(BeNil())
	})

	It("does not detect meshes that haven't been assigned an ARN", func() {
		awsMesh := aws_v1beta2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:        meshName,
				ClusterName: clusterName,
			},
			Spec: aws_v1beta2.MeshSpec{
				AWSName: &meshName,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"mesh": meshName},
				},
			},
			Status: aws_v1beta2.MeshStatus{},
		}

		builder := input.NewInputSnapshotManualBuilder("app mesh test")
		builder.AddMeshes([]*aws_v1beta2.Mesh{&awsMesh})

		actual, err := meshDetector.DetectMeshes(builder.Build())
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(BeNil())
	})

	It("errors when an ARN is malformed", func() {
		badArn := "this can't be parsed as a Mesh ARN"

		awsMesh := aws_v1beta2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:        meshName,
				ClusterName: clusterName,
			},
			Spec: aws_v1beta2.MeshSpec{
				AWSName: &meshName,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"mesh": meshName},
				},
			},
			Status: aws_v1beta2.MeshStatus{
				MeshARN: &badArn,
			},
		}

		builder := input.NewInputSnapshotManualBuilder("app mesh test")
		builder.AddMeshes([]*aws_v1beta2.Mesh{&awsMesh})

		_, err := meshDetector.DetectMeshes(builder.Build())
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(appmesh.UnexpectedMeshARNError(badArn)))
	})

})
