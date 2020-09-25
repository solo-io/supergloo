package detector_test

import (
	"context"

	"github.com/solo-io/skv2/pkg/ezkube"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector"
	mock_detector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("WorkloadDetector", func() {

	var (
		ctrl                *gomock.Controller
		mockSidecarDetector *mock_detector.MockSidecarDetector
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockSidecarDetector = mock_detector.NewMockSidecarDetector(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	deploymentName := "name"
	deploymentNs := "namespace"
	deploymentCluster := "cluster"
	serviceAccountName := "service-account-name"
	podLabels := map[string]string{"a": "b"}

	makeDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   deploymentNs,
				ClusterName: deploymentCluster,
				Name:        deploymentName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: podLabels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "istio-pilot:latest",
							},
						},
						ServiceAccountName: serviceAccountName,
					},
				},
			},
		}
	}

	makeReplicaSet := func(dep *appsv1.Deployment) *appsv1.ReplicaSet {
		rs := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   deploymentNs,
				ClusterName: deploymentCluster,
				Name:        "replicaset",
			},
		}
		err := controllerutil.SetControllerReference(dep, rs, scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())
		return rs
	}

	makePod := func(rs *appsv1.ReplicaSet) *corev1.Pod {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   deploymentNs,
				Name:        "pod",
				ClusterName: deploymentCluster,
			},
		}
		err := controllerutil.SetControllerReference(rs, pod, scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())
		return pod
	}

	mesh := &v1alpha2.Mesh{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mesh",
			Namespace: "service-mesh-hub",
		},
	}

	It("translates a deployment with a detected sidecar to a workload", func() {

		deployment := makeDeployment()
		rs := makeReplicaSet(deployment)
		pod := makePod(rs)

		pods := corev1sets.NewPodSet(pod)
		replicaSets := appsv1sets.NewReplicaSetSet(rs)
		detector := NewWorkloadDetector(
			context.TODO(),
			pods,
			replicaSets,
			mockSidecarDetector,
		)

		meshes := v1alpha2sets.NewMeshSet()

		mockSidecarDetector.EXPECT().DetectMeshSidecar(pod, meshes).Return(mesh)

		workload := detector.DetectWorkload(types.ToWorkload(deployment), meshes)

		outputMeta := utils.DiscoveredObjectMeta(deployment)
		// expect appended workload kind
		outputMeta.Name += "-deployment"

		Expect(workload).To(Equal(&v1alpha2.Workload{
			ObjectMeta: outputMeta,
			Spec: v1alpha2.WorkloadSpec{
				WorkloadType: &v1alpha2.WorkloadSpec_Kubernetes{
					Kubernetes: &v1alpha2.WorkloadSpec_KubernetesWorkload{
						Controller:         ezkube.MakeClusterObjectRef(deployment),
						PodLabels:          podLabels,
						ServiceAccountName: serviceAccountName,
					},
				},
				Mesh: ezkube.MakeObjectRef(mesh),
			},
		}))
	})

})
