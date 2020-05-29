package appmesh_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/appmesh"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	appsv1 "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AppmeshMeshWorkloadScanner", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		mockOwnerFetcher        *mock_mesh_workload.MockOwnerFetcher
		mockMeshClient          *mock_core.MockMeshClient
		mockAppMeshParser       *mock_aws.MockAppMeshScanner
		mockRemoteClient        *mock_controller_runtime.MockClient
		mockAwsAccountIdFetcher *mock_aws.MockAwsAccountIdFetcher
		meshWorkloadScanner     k8s.MeshWorkloadScanner
		namespace               = "namespace"
		clusterName             = "clusterName"
		deploymentName          = "deployment-name"
		deploymentKind          = "deployment-kind"
		meshName                = "mesh-name-1"
		region                  = "us-east-1"
		awsAccountId            = "awsaccountid"
		ports                   = []*zephyr_discovery_types.MeshWorkloadSpec_Appmesh_ContainerPort{
			{
				Port:     9080,
				Protocol: string(k8s_core_types.ProtocolTCP),
			},
			{
				Port:     9081,
				Protocol: string(k8s_core_types.ProtocolUDP),
			},
			{
				Port:     9082,
				Protocol: string(k8s_core_types.ProtocolTCP),
			},
		}
		pod = &k8s_core_types.Pod{
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{
						Image: "appmesh-envoy",
						Env: []k8s_core_types.EnvVar{
							{
								Name:  aws_utils.AppMeshVirtualNodeEnvVarName,
								Value: fmt.Sprintf("mesh/%s/virtualNode/virtualNodeName", meshName),
							},
							{
								Name:  aws_utils.AppMeshRegionEnvVarName,
								Value: region,
							},
							{
								Name:  aws_utils.AppMeshRoleArnEnvVarName,
								Value: fmt.Sprintf("arn:aws:iam::%s:role/iamserviceaccount-role", awsAccountId),
							},
						},
					},
				},
			},
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace},
		}
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
			TypeMeta:   metav1.TypeMeta{Kind: deploymentKind},
			Spec: appsv1.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
							{
								Ports: []k8s_core_types.ContainerPort{
									{
										ContainerPort: int32(ports[0].Port),
										Protocol:      k8s_core_types.Protocol(ports[0].Protocol),
									},
									{
										ContainerPort: int32(ports[1].Port),
										Protocol:      k8s_core_types.Protocol(ports[1].Protocol),
									},
								},
							},
							{
								Ports: []k8s_core_types.ContainerPort{
									{
										ContainerPort: int32(ports[2].Port),
										Protocol:      k8s_core_types.Protocol(ports[2].Protocol),
									},
								},
							},
						},
					},
				},
			},
		}
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockOwnerFetcher = mock_mesh_workload.NewMockOwnerFetcher(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockAppMeshParser = mock_aws.NewMockAppMeshScanner(ctrl)
		mockRemoteClient = mock_controller_runtime.NewMockClient(ctrl)
		mockAwsAccountIdFetcher = mock_aws.NewMockAwsAccountIdFetcher(ctrl)
		meshWorkloadScanner = appmesh.NewAppMeshWorkloadScanner(
			mockOwnerFetcher,
			mockAppMeshParser,
			mockMeshClient,
			mockAwsAccountIdFetcher,
			mockRemoteClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should scan pod and disambiguate multiple AppMesh Meshes", func() {
		virtualNodeName := "virtualNodeName"
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(deployment, nil)
		mesh := &zephyr_discovery.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meshName,
				Namespace: env.GetWriteNamespace(),
			},
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
				MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
						Name:   meshName,
						Region: region,
					},
				},
			},
		}
		mockMeshClient.
			EXPECT().
			GetMesh(
				ctx,
				client.ObjectKey{
					Name:      metadata.BuildAppMeshName(meshName, region, awsAccountId),
					Namespace: env.GetWriteNamespace(),
				},
			).Return(mesh, nil)
		expectedMeshWorkload := &zephyr_discovery.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-%s-%s", "appmesh", deploymentName, namespace, clusterName),
				Namespace: env.GetWriteNamespace(),
				Labels:    appmesh.DiscoveryLabels(),
			},
			Spec: zephyr_discovery_types.MeshWorkloadSpec{
				KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &zephyr_core_types.ResourceRef{
						Name:      deployment.Name,
						Namespace: deployment.Namespace,
						Cluster:   clusterName,
					},
					Labels:             nil,
					ServiceAccountName: "",
				},
				Mesh: &zephyr_core_types.ResourceRef{
					Name:      meshName,
					Namespace: mesh.GetNamespace(),
				},
				Appmesh: &zephyr_discovery_types.MeshWorkloadSpec_Appmesh{
					VirtualNodeName: virtualNodeName,
					Ports:           ports,
				},
			},
		}
		mockAwsAccountIdFetcher.EXPECT().GetEksAccountId(ctx, mockRemoteClient).Return(aws_utils.AwsAccountId(awsAccountId), nil)
		mockAppMeshParser.
			EXPECT().
			ScanPodForAppMesh(pod, aws_utils.AwsAccountId(awsAccountId)).
			Return(&aws_utils.AppMeshPod{
				Region:          region,
				AppMeshName:     meshName,
				AwsAccountID:    awsAccountId,
				VirtualNodeName: virtualNodeName,
			}, nil)
		meshWorkload, err := meshWorkloadScanner.ScanPod(ctx, pod, clusterName)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWorkload).To(Equal(expectedMeshWorkload))
	})

	It("should return nil if not appmesh injected pod", func() {
		nonAppMeshPod := &k8s_core_types.Pod{
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{Image: "random-image"},
				},
			},
			ObjectMeta: metav1.ObjectMeta{ClusterName: clusterName, Namespace: namespace},
		}
		mockAwsAccountIdFetcher.EXPECT().GetEksAccountId(ctx, mockRemoteClient).Return(aws_utils.AwsAccountId(""), nil)
		mockAppMeshParser.
			EXPECT().
			ScanPodForAppMesh(nonAppMeshPod, aws_utils.AwsAccountId("")).
			Return(nil, nil)
		meshWorkload, err := meshWorkloadScanner.ScanPod(ctx, nonAppMeshPod, clusterName)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWorkload).To(BeNil())
	})

	It("should return error if error fetching deployment", func() {
		expectedErr := eris.New("error")
		mockAwsAccountIdFetcher.EXPECT().GetEksAccountId(ctx, mockRemoteClient).Return(aws_utils.AwsAccountId(""), nil)
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(nil, expectedErr)
		mockAppMeshParser.
			EXPECT().
			ScanPodForAppMesh(pod, aws_utils.AwsAccountId("")).
			Return(&aws_utils.AppMeshPod{}, nil)
		_, err := meshWorkloadScanner.ScanPod(ctx, pod, clusterName)
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
	})
})
