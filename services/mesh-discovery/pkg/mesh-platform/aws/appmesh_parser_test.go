package aws_test

import (
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws/mocks"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AppmeshParser", func() {
	var (
		ctrl          *gomock.Controller
		mockArnParser *mock_aws.MockArnParser
		appMeshParser aws_utils.AppMeshParser
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		appMeshParser = aws_utils.NewAppMeshParser(mockArnParser)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should scan pod for AppMesh sidecar and return data if so", func() {
		expectedAppMeshPod := &aws_utils.AppMeshPod{
			AwsAccountID:    "account-id",
			Region:          "us-east-2",
			AppMeshName:     "appmeshname",
			VirtualNodeName: "virtualnodename",
		}
		pod := &k8s_core_types.Pod{
			ObjectMeta: k8s_meta_types.ObjectMeta{},
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{
						Image: "840364872350.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.12.2.1-prod",
						Env: []k8s_core_types.EnvVar{
							{
								Name: aws_utils.AppMeshVirtualNodeEnvVarName,
								Value: fmt.Sprintf(
									"mesh/%s/virtualNode/%s",
									expectedAppMeshPod.AppMeshName,
									expectedAppMeshPod.VirtualNodeName,
								),
							},
							{
								Name:  aws_utils.AppMeshRegionEnvVarName,
								Value: expectedAppMeshPod.Region,
							},
							{
								Name:  aws_utils.AppMeshRoleArnEnvVarName,
								Value: fmt.Sprintf("arn:aws:iam::%s:role/iamserviceaccount-role", expectedAppMeshPod.AwsAccountID),
							},
						},
					},
				},
			},
		}
		mockArnParser.EXPECT().ParseAccountID(pod.Spec.Containers[0].Env[2].Value).Return(expectedAppMeshPod.AwsAccountID, nil)
		appMeshPod, err := appMeshParser.ScanPodForAppMesh(pod)
		Expect(err).To(BeNil())
		Expect(appMeshPod).To(Equal(expectedAppMeshPod))
	})

	It("should return error if empty env var value", func() {
		pod := &k8s_core_types.Pod{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{
						Image: "840364872350.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.12.2.1-prod",
						Env: []k8s_core_types.EnvVar{
							{
								Name:  aws_utils.AppMeshVirtualNodeEnvVarName,
								Value: "mesh/foo/virtualNode/bar",
							},
							{
								Name:  aws_utils.AppMeshRegionEnvVarName,
								Value: "",
							},
							{
								Name:  aws_utils.AppMeshRoleArnEnvVarName,
								Value: "foobar",
							},
						},
					},
				},
			},
		}
		_, err := appMeshParser.ScanPodForAppMesh(pod)
		Expect(err).To(testutils.HaveInErrorChain(aws_utils.EmptyEnvVarValueError(aws_utils.AppMeshRegionEnvVarName, pod.ObjectMeta)))
	})

	It("should return error if virtualnode env var has unexpected format", func() {
		unexpectedValue := "incorrect format"
		pod := &k8s_core_types.Pod{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{
						Image: "840364872350.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.12.2.1-prod",
						Env: []k8s_core_types.EnvVar{
							{
								Name:  aws_utils.AppMeshVirtualNodeEnvVarName,
								Value: unexpectedValue,
							},
							{
								Name:  aws_utils.AppMeshRegionEnvVarName,
								Value: "asdf",
							},
							{
								Name:  aws_utils.AppMeshRoleArnEnvVarName,
								Value: "foobar",
							},
						},
					},
				},
			},
		}
		_, err := appMeshParser.ScanPodForAppMesh(pod)
		Expect(err).To(testutils.HaveInErrorChain(aws_utils.UnexpectedVirtualNodeValue(unexpectedValue)))
	})

	It("should return nil if not AppMesh injected", func() {
		pod := &k8s_core_types.Pod{
			ObjectMeta: k8s_meta_types.ObjectMeta{},
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{
						Image: "some-rando-image",
					},
				},
			},
		}
		appMeshPod, err := appMeshParser.ScanPodForAppMesh(pod)
		Expect(err).To(BeNil())
		Expect(appMeshPod).To(BeNil())
	})
})
