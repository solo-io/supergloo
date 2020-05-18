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
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/appmesh"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AppmeshMeshWorkloadScanner", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		mockOwnerFetcher    *mock_mesh_workload.MockOwnerFetcher
		mockMeshClient      *mock_core.MockMeshClient
		mockAppMeshParser   *mock_aws.MockAppMeshScanner
		mockConfigMapClient *mock_kubernetes_core.MockConfigMapClient
		meshWorkloadScanner k8s.MeshWorkloadScanner
		namespace           = "namespace"
		clusterName         = "clusterName"
		deploymentName      = "deployment-name"
		deploymentKind      = "deployment-kind"
		meshName            = "mesh-name-1"
		region              = "us-east-1"
		awsAccountId        = "awsaccountid"
		pod                 = &k8s_core_types.Pod{
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
		}
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockOwnerFetcher = mock_mesh_workload.NewMockOwnerFetcher(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockAppMeshParser = mock_aws.NewMockAppMeshScanner(ctrl)
		mockConfigMapClient = mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		meshWorkloadScanner = appmesh.NewAppMeshWorkloadScanner(
			mockOwnerFetcher,
			mockAppMeshParser,
			mockMeshClient,
			func(client client.Client) k8s_core.ConfigMapClient {
				return mockConfigMapClient
			},
			nil,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should scan pod and disambiguate multiple AppMesh Meshes", func() {
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
			},
		}
		configMap := &k8s_core_types.ConfigMap{}
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, appmesh.AwsAuthConfigMapKey).Return(configMap, nil)
		mockAppMeshParser.
			EXPECT().
			ScanPodForAppMesh(pod, configMap).
			Return(&aws_utils.AppMeshPod{
				Region:       region,
				AppMeshName:  meshName,
				AwsAccountID: awsAccountId,
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
		configMap := &k8s_core_types.ConfigMap{}
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, appmesh.AwsAuthConfigMapKey).Return(configMap, nil)
		mockAppMeshParser.
			EXPECT().
			ScanPodForAppMesh(nonAppMeshPod, configMap).
			Return(nil, nil)
		meshWorkload, err := meshWorkloadScanner.ScanPod(ctx, nonAppMeshPod, clusterName)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWorkload).To(BeNil())
	})

	It("should return error if error fetching deployment", func() {
		expectedErr := eris.New("error")
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(nil, expectedErr)
		configMap := &k8s_core_types.ConfigMap{}
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, appmesh.AwsAuthConfigMapKey).Return(configMap, nil)
		mockAppMeshParser.
			EXPECT().
			ScanPodForAppMesh(pod, configMap).
			Return(&aws_utils.AppMeshPod{}, nil)
		_, err := meshWorkloadScanner.ScanPod(ctx, pod, clusterName)
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
	})
})
