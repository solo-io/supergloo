package gloo

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	kubeClient kubernetes.Interface
)

var _ = Describe("gloo registration syncers", func() {
	var (
		defaultMode  int32 = 420
		optional           = true
		meshResource       = &core.ResourceRef{
			Name:      "istio",
			Namespace: "istio-system",
		}
		istioMesh = &v1.Mesh{
			MtlsConfig: &v1.MtlsConfig{
				MtlsEnabled: true,
			},
			MeshType: &v1.Mesh_Istio{
				Istio: &v1.IstioMesh{
					InstallationNamespace: "istio-system",
				},
			},
			Metadata: core.Metadata{
				Name:      "istio",
				Namespace: "istio-system",
			},
		}
		volumeList = VolumeList{
			corev1.Volume{
				Name: certVolumeName(meshResource),
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "istio.default",
						Items:       nil,
						DefaultMode: &defaultMode,
						Optional:    &optional,
					},
				},
			},
			corev1.Volume{
				Name: "1",
			},
			corev1.Volume{
				Name: "2",
			},
		}
		mountList = VolumeMountList{
			corev1.VolumeMount{
				Name:      certVolumeName(meshResource),
				ReadOnly:  true,
				MountPath: certVolumePathName(meshResource),
			},
			corev1.VolumeMount{
				Name: "1",
			},
			corev1.VolumeMount{
				Name: "2",
			},
		}
		glooIngress = func(meshes ...*core.ResourceRef) *v1.MeshIngress {
			return &v1.MeshIngress{
				InstallationNamespace: "gloo-system",
				MeshIngressType: &v1.MeshIngress_Gloo{
					Gloo: &v1.GlooMeshIngress{},
				},
				Meshes: meshes,
			}
		}
	)
	var createDeployment = func(volumes VolumeList, mounts VolumeMountList) *v1beta1.Deployment {
		return &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gateway-proxy",
				Namespace: "gloo-system",
			},
			Spec: v1beta1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Volumes: volumes,
						Containers: []corev1.Container{
							{
								VolumeMounts: mounts,
							},
						},
					},
				},
			},
		}
	}
	Context("should update", func() {
		It("Should not update if nothing changed", func() {
			deployment := createDeployment(volumeList, mountList)
			update, err := shouldUpdateDeployment(deployment, []*core.ResourceRef{meshResource}, v1.MeshList{istioMesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeFalse())
			Expect(deployment).To(Equal(createDeployment(volumeList, mountList)))
		})

		It("should update if one is added", func() {
			deployment := createDeployment(volumeList.Remove(0), mountList.Remove(0))
			update, err := shouldUpdateDeployment(deployment, []*core.ResourceRef{meshResource}, v1.MeshList{istioMesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeTrue())
			Expect(deployment).NotTo(Equal(createDeployment(volumeList.Remove(0), mountList.Remove(0))))
		})

		It("should update if one is removed", func() {
			deployment := createDeployment(volumeList, mountList)
			update, err := shouldUpdateDeployment(deployment, []*core.ResourceRef{}, v1.MeshList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeTrue())
			Expect(deployment).NotTo(Equal(createDeployment(volumeList, mountList)))
		})

		It("should update if mtls is enabled", func() {
			deployment := createDeployment(volumeList.Remove(0), mountList.Remove(0))
			update, err := shouldUpdateDeployment(deployment, []*core.ResourceRef{meshResource}, v1.MeshList{istioMesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeTrue())
			Expect(deployment).NotTo(Equal(createDeployment(volumeList, mountList)))
		})

		It("should update if mtls is disabled", func() {
			deployment := createDeployment(volumeList, mountList)
			newMesh := &v1.Mesh{
				MtlsConfig: &v1.MtlsConfig{
					MtlsEnabled: false,
				},
				Metadata: istioMesh.Metadata,
				MeshType: istioMesh.MeshType,
				Status:   istioMesh.Status,
			}
			update, err := shouldUpdateDeployment(deployment, []*core.ResourceRef{meshResource}, v1.MeshList{newMesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeTrue())
			Expect(deployment).NotTo(Equal(createDeployment(volumeList, mountList)))
		})
	})

	Context("update deployments properly", func() {
		var (
			syncer v1.RegistrationSyncer
			cs     *clientset.Clientset
			ctx    context.Context
		)

		BeforeEach(func() {
			var err error
			ctx = context.Background()
			cs, err = clientset.ClientsetFromContext(ctx)
			Expect(err).NotTo(HaveOccurred())
			cs.Kube = kubeClient
			newReporter := reporter.NewReporter("gloo-registration-reporter",
				cs.Supergloo.Mesh.BaseClient(),
				cs.Supergloo.MeshIngress.BaseClient())
			syncer = NewGlooRegistrationSyncer(newReporter, cs)
		})

		var checkDeployment = func(volumes VolumeList, mounts VolumeMountList) {
			deployment, err := kubeClient.ExtensionsV1beta1().Deployments("gloo-system").Get("gateway-proxy", metav1.GetOptions{})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, len(deployment.Spec.Template.Spec.Containers)).To(BeNumerically(">", 0))
			gatewayProxyContainer := deployment.Spec.Template.Spec.Containers[0]
			for _, v := range mounts {
				ExpectWithOffset(1, gatewayProxyContainer.VolumeMounts).To(ContainElement(v))
			}
			for _, v := range volumes {
				ExpectWithOffset(1, deployment.Spec.Template.Spec.Volumes).To(ContainElement(v))
			}
		}

		AfterEach(func() {
			err := kubeClient.ExtensionsV1beta1().Deployments("gloo-system").Delete("gateway-proxy", &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("does nothing when states haven't changed", func() {
			_, err := kubeClient.ExtensionsV1beta1().Deployments("gloo-system").Create(createDeployment(volumeList, mountList))
			Expect(err).NotTo(HaveOccurred())
			snap := &v1.RegistrationSnapshot{
				Meshes: v1.MeshesByNamespace{
					"istio-system": v1.MeshList{istioMesh},
				},
				Meshingresses: v1.MeshingressesByNamespace{
					"gloo-system": v1.MeshIngressList{glooIngress(meshResource)},
				},
			}
			err = syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			checkDeployment(volumeList, mountList)
		})
		It("Adds missing volume", func() {
			_, err := kubeClient.ExtensionsV1beta1().Deployments("gloo-system").Create(createDeployment(volumeList.Remove(0), mountList.Remove(0)))
			Expect(err).NotTo(HaveOccurred())
			snap := &v1.RegistrationSnapshot{
				Meshes: v1.MeshesByNamespace{
					"istio-system": v1.MeshList{istioMesh},
				},
				Meshingresses: v1.MeshingressesByNamespace{
					"gloo-system": v1.MeshIngressList{glooIngress(meshResource)},
				},
			}
			err = syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			checkDeployment(volumeList, mountList)
		})
		It("deletes extra volume", func() {
			_, err := kubeClient.ExtensionsV1beta1().Deployments("gloo-system").Create(createDeployment(volumeList, mountList))
			Expect(err).NotTo(HaveOccurred())
			snap := &v1.RegistrationSnapshot{
				Meshes: v1.MeshesByNamespace{
					"istio-system": v1.MeshList{istioMesh},
				},
				Meshingresses: v1.MeshingressesByNamespace{
					"gloo-system": v1.MeshIngressList{glooIngress()},
				},
			}
			err = syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			checkDeployment(volumeList.Remove(0), mountList.Remove(0))
		})
		It("errors when mesh ins't available", func() {
			_, err := kubeClient.ExtensionsV1beta1().Deployments("gloo-system").Create(createDeployment(volumeList, mountList))
			Expect(err).NotTo(HaveOccurred())
			snap := &v1.RegistrationSnapshot{
				Meshes: v1.MeshesByNamespace{},
				Meshingresses: v1.MeshingressesByNamespace{
					"gloo-system": v1.MeshIngressList{glooIngress(meshResource)},
				},
			}
			err = syncer.Sync(ctx, snap)
			Expect(err).To(HaveOccurred())
		})
	})

})
