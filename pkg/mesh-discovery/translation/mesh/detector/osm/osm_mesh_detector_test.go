package osm_test

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/osm"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("OsmMeshDetector", func() {

	ctx := context.Background()
	serviceAccountName := "service-account-name"
	meshNs := "namespace"
	clusterName := "cluster"

	osmController := func(deploymentName string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   meshNs,
				Name:        deploymentName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "osm-controller:latest",
							},
						},
						ServiceAccountName: serviceAccountName,
					},
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "osm"},
				},
			},
		}
	}

	It("does not detect OSM when it is not there", func() {

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: "a", Name: "a"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "test-image",
							},
						},
					},
				},
			},
		}

		detector := NewMeshDetector(
			ctx,
		)

		in := input.NewInputSnapshotManualBuilder("")
		in.AddDeployments([]*appsv1.Deployment{deployment})

		meshes, err := detector.DetectMeshes(in.Build())
		Expect(err).NotTo(HaveOccurred())
		Expect(meshes).To(BeNil())
	})

	It("detects a mesh from a deployment named osm-controller", func() {

		detector := NewMeshDetector(
			ctx,
		)

		deployment := osmController("osm-controller")
		expected := &v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "osm-controller-namespace-cluster",
				Namespace: defaults.GetPodNamespace(),
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: v1alpha2.MeshSpec{
				MeshType: &v1alpha2.MeshSpec_Osm{
					Osm: &v1alpha2.MeshSpec_OSM{
						Installation: &v1alpha2.MeshSpec_MeshInstallation{
							Namespace: meshNs,
							Cluster:   clusterName,
							Version:   "latest",
							PodLabels: map[string]string{"app": "osm"},
						},
					},
				},
			},
		}

		in := input.NewInputSnapshotManualBuilder("")
		in.AddDeployments([]*appsv1.Deployment{deployment})

		meshes, err := detector.DetectMeshes(in.Build())
		Expect(err).NotTo(HaveOccurred())
		Expect(meshes).To(HaveLen(1))
		Expect(meshes[0]).To(Equal(expected))
	})

})
