package appmesh

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector"
)

var _ = Describe("AppMesh SidecarDetector", func() {

	var (
		sidecarDetector detector.SidecarDetector
	)

	BeforeEach(func() {
		sidecarDetector = NewSidecarDetector(context.Background())
	})

	It("returns the corresponding mesh for an appmesh-injected pod", func() {
		meshName := "mesh-name"
		virtualNodeName := fmt.Sprintf("mesh/%s/virtualNode/<virtual-node-name>", meshName)
		injectedPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "name",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "application-container",
						Image: "quay.io/solo-io/app-container:1.0.0",
					},
					{
						Name:  "sidecar-container",
						Image: "123456.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.15.0.0-prod",
						Env: []corev1.EnvVar{
							{
								Name:  "foo",
								Value: "bar",
							},
							{
								Name:  appMeshVirtualNodeEnvVarName,
								Value: virtualNodeName,
							},
						},
					},
				},
			},
		}

		mesh := &v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", meshName, "some-cluster"),
				Namespace: "service-mesh-hub",
			},
			Spec: v1alpha2.MeshSpec{
				MeshType: &v1alpha2.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &v1alpha2.MeshSpec_AwsAppMesh{
						AwsName: meshName,
					},
				},
			},
		}

		actual := sidecarDetector.DetectMeshSidecar(injectedPod, v1alpha2sets.NewMeshSet(mesh))
		Expect(actual).To(Equal(mesh))
	})

	It("returns nil for a pod with no sidecar", func() {
		meshName := "mesh-name"
		plainPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "name",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "application-container",
						Image: "quay.io/solo-io/app-container:1.0.0",
					},
				},
			},
		}

		mesh := &v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", meshName, "some-cluster"),
				Namespace: "service-mesh-hub",
			},
			Spec: v1alpha2.MeshSpec{
				MeshType: &v1alpha2.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &v1alpha2.MeshSpec_AwsAppMesh{
						AwsName: meshName,
					},
				},
			},
		}

		actual := sidecarDetector.DetectMeshSidecar(plainPod, v1alpha2sets.NewMeshSet(mesh))
		Expect(actual).To(BeNil())
	})

	It("returns nil when a corresponding mesh cannot be found for the sidecar", func() {
		meshName := "mesh-name"
		virtualNodeName := fmt.Sprintf("mesh/%s/virtualNode/<virtual-node-name>", meshName)
		injectedPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "name",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "application-container",
						Image: "quay.io/solo-io/app-container:1.0.0",
					},
					{
						Name:  "sidecar-container",
						Image: "123456.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.15.0.0-prod",
						Env: []corev1.EnvVar{
							{
								Name:  "foo",
								Value: "bar",
							},
							{
								Name:  appMeshVirtualNodeEnvVarName,
								Value: virtualNodeName,
							},
						},
					},
				},
			},
		}

		mesh := &v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", meshName, "some-cluster"),
				Namespace: "service-mesh-hub",
			},
			Spec: v1alpha2.MeshSpec{
				MeshType: &v1alpha2.MeshSpec_Istio_{},
			},
		}

		actual := sidecarDetector.DetectMeshSidecar(injectedPod, v1alpha2sets.NewMeshSet(mesh))
		Expect(actual).To(BeNil())
	})

	It("returns nil when the VirtualNodeName is malformed", func() {
		meshName := "mesh-name"
		virtualNodeName := "invalid"
		injectedPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "name",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "application-container",
						Image: "quay.io/solo-io/app-container:1.0.0",
					},
					{
						Name:  "sidecar-container",
						Image: "123456.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.15.0.0-prod",
						Env: []corev1.EnvVar{
							{
								Name:  "foo",
								Value: "bar",
							},
							{
								Name:  appMeshVirtualNodeEnvVarName,
								Value: virtualNodeName,
							},
						},
					},
				},
			},
		}

		mesh := &v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", meshName, "some-cluster"),
				Namespace: "service-mesh-hub",
			},
			Spec: v1alpha2.MeshSpec{
				MeshType: &v1alpha2.MeshSpec_Istio_{},
			},
		}

		actual := sidecarDetector.DetectMeshSidecar(injectedPod, v1alpha2sets.NewMeshSet(mesh))
		Expect(actual).To(BeNil())
	})

})
