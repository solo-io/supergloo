package appmesh

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	glooinstall "github.com/solo-io/gloo/pkg/cliutil/install"
	gloomocks "github.com/solo-io/gloo/projects/gloo/pkg/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	cliclients "github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	climocks "github.com/solo-io/supergloo/cli/pkg/helpers/mocks"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Syncer", func() {

	var (
		kube      kubernetes.Interface
		ctrl      *gomock.Controller
		syncer    v1.RegistrationSyncer
		installer mockInstaller
		ctx       context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		defer ctrl.Finish()

		kube = fake.NewSimpleClientset()

		// Mock client used to get AWS secret
		secretClient := gloomocks.NewMockSecretClient(ctrl)
		secretClient.EXPECT().Read(testNamespace, "aws-secret", gomock.Any()).Return(awsSecret, nil).AnyTimes()
		secretClient.EXPECT().Read(testNamespace, gomock.Not("aws-secret"), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()

		ctx = context.Background()

		// Create the config map containing the manifests for the auto-injection resources
		autoInjectionResourcesConfigMap := &corev1.ConfigMap{}
		cm, err := ioutil.ReadFile("test/test-sidecar-injector-configmap.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(yaml.Unmarshal(cm, autoInjectionResourcesConfigMap)).NotTo(HaveOccurred())
		_, err = kube.CoreV1().ConfigMaps(testNamespace).Create(autoInjectionResourcesConfigMap)
		Expect(err).NotTo(HaveOccurred())

		// Mock AWS client
		successMock := climocks.NewMockAppmesh(ctrl)
		successMock.EXPECT().ListMeshes(nil).Return(&appmesh.ListMeshesOutput{}, nil).AnyTimes()
		cliclients.UseAppmeshMock(successMock)

		installer = mockInstaller{}

		syncer = NewAppMeshRegistrationSyncer(mockReporter{}, kube, secretClient, &installer)
	})

	AfterEach(func() {
		Expect(kube.CoreV1().ConfigMaps(testNamespace).Delete(resourcesConfigMapName, &metav1.DeleteOptions{})).NotTo(HaveOccurred())
	})

	It("remove auto-injection resources if snapshot does not contain App Mesh meshes", func() {
		err := syncer.Sync(ctx, &v1.RegistrationSnapshot{
			Meshes: v1.MeshesByNamespace{
				testNamespace: v1.MeshList{istio},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(installer.Created).To(HaveLen(0))
		Expect(installer.Deleted).To(HaveLen(5))
		Expect(installer.Exists(secretKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(deploymentKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(serviceKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(configMapKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(webhookConfigKind, "", webhookName)).To(BeFalse())
	})

	It("remove auto-injection resources if auto-injection is disabled for all meshes", func() {
		err := syncer.Sync(ctx, &v1.RegistrationSnapshot{
			Meshes: v1.MeshesByNamespace{
				testNamespace: v1.MeshList{
					istio,
					appMeshNoAutoInjection,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(installer.Created).To(HaveLen(0))
		Expect(installer.Deleted).To(HaveLen(5))
		Expect(installer.Exists(secretKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(deploymentKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(serviceKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(configMapKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(webhookConfigKind, "", webhookName)).To(BeFalse())
	})

	It("creates auto-injection resources if snapshot contains App Mesh mesh", func() {
		err := syncer.Sync(ctx, &v1.RegistrationSnapshot{
			Meshes: v1.MeshesByNamespace{
				testNamespace: v1.MeshList{mesh},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(installer.Created).To(HaveLen(5))
		Expect(installer.Deleted).To(HaveLen(0))
		Expect(installer.Exists(secretKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(deploymentKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(serviceKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(configMapKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(webhookConfigKind, "", webhookName)).To(BeTrue())
	})

	It("does nothing if all auto-injection resources are already in place", func() {
		_, err := kube.CoreV1().Secrets(testNamespace).Create(secret)
		Expect(err).NotTo(HaveOccurred())
		_, err = kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(whConfig)
		Expect(err).NotTo(HaveOccurred())
		_, err = kube.AppsV1().Deployments(testNamespace).Create(deployment)
		Expect(err).NotTo(HaveOccurred())
		_, err = kube.CoreV1().Services(testNamespace).Create(service)
		Expect(err).NotTo(HaveOccurred())
		_, err = kube.CoreV1().ConfigMaps(testNamespace).Create(configMap)
		Expect(err).NotTo(HaveOccurred())

		err = syncer.Sync(ctx, &v1.RegistrationSnapshot{
			Meshes: v1.MeshesByNamespace{
				testNamespace: v1.MeshList{mesh},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(installer.Created).To(HaveLen(0))
		Expect(installer.Deleted).To(HaveLen(0))
		Expect(installer.Exists(secretKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(deploymentKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(serviceKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(configMapKind, testNamespace, webhookName)).To(BeFalse())
		Expect(installer.Exists(webhookConfigKind, "", webhookName)).To(BeFalse())
	})

	It("recreates all auto-injection resources if they all are in an inconsistent state", func() {
		_, err := kube.CoreV1().Secrets(testNamespace).Create(secret)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(secretKind, testNamespace, webhookName)

		_, err = kube.AppsV1().Deployments(testNamespace).Create(deployment)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(deploymentKind, testNamespace, webhookName)

		err = syncer.Sync(ctx, &v1.RegistrationSnapshot{
			Meshes: v1.MeshesByNamespace{
				testNamespace: v1.MeshList{mesh},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(installer.Created).To(HaveLen(5))
		Expect(installer.Deleted).To(HaveLen(5))
		Expect(installer.Exists(secretKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(deploymentKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(serviceKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(configMapKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(webhookConfigKind, "", webhookName)).To(BeTrue())
	})

	It("recreates only secret auto-injection resources if they are in an inconsistent state", func() {
		_, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(whConfig)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(webhookConfigKind, "", webhookName)

		_, err = kube.AppsV1().Deployments(testNamespace).Create(deployment)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(deploymentKind, testNamespace, webhookName)

		_, err = kube.CoreV1().Services(testNamespace).Create(service)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(serviceKind, testNamespace, webhookName)

		_, err = kube.CoreV1().ConfigMaps(testNamespace).Create(configMap)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(configMapKind, testNamespace, webhookName)

		err = syncer.Sync(ctx, &v1.RegistrationSnapshot{
			Meshes: v1.MeshesByNamespace{
				testNamespace: v1.MeshList{mesh},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(installer.Created).To(HaveLen(2))
		Expect(installer.Deleted).To(HaveLen(2))
		Expect(installer.Exists(secretKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(deploymentKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(serviceKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(configMapKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(webhookConfigKind, "", webhookName)).To(BeTrue())
	})

	It("recreates only non-secret auto-injection resources if they are in an inconsistent state", func() {
		_, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(whConfig)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(webhookConfigKind, "", webhookName)

		_, err = kube.CoreV1().Secrets(testNamespace).Create(secret)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(secretKind, testNamespace, webhookName)

		_, err = kube.AppsV1().Deployments(testNamespace).Create(deployment)
		Expect(err).NotTo(HaveOccurred())
		installer.Add(deploymentKind, testNamespace, webhookName)

		err = syncer.Sync(ctx, &v1.RegistrationSnapshot{
			Meshes: v1.MeshesByNamespace{
				testNamespace: v1.MeshList{mesh},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(installer.Created).To(HaveLen(3))
		Expect(installer.Deleted).To(HaveLen(3))
		Expect(installer.Exists(secretKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(deploymentKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(serviceKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(configMapKind, testNamespace, webhookName)).To(BeTrue())
		Expect(installer.Exists(webhookConfigKind, "", webhookName)).To(BeTrue())
	})
})

type mockReporter struct{}

func (mockReporter) WriteReports(ctx context.Context, errs reporter.ResourceErrors, subresourceStatuses map[string]*core.Status) error {
	return nil
}

type mockInstaller struct {
	Created, Deleted, Existing []glooinstall.ResourceType
}

func (m *mockInstaller) Delete(namespace string, reader io.Reader) error {
	resources, err := m.getResources(reader)
	if err != nil {
		return err
	}
	m.Deleted = append(m.Deleted, resources...)

	// Remove from existing
	for i, existing := range m.Existing {
		for _, deleted := range resources {
			if deleted.Kind == existing.Kind && deleted.Metadata.Namespace == existing.Metadata.Namespace && deleted.Metadata.Name == existing.Metadata.Name {
				m.Existing = append(m.Existing[:i], m.Existing[i+1:]...)
			}
		}
	}

	return nil
}

func (m *mockInstaller) Create(namespace string, reader io.Reader, timeout int64, shouldWait bool) error {
	resources, err := m.getResources(reader)
	if err != nil {
		return err
	}
	m.Created = append(m.Created, resources...)
	m.Existing = append(m.Existing, resources...)
	return nil
}

func (m *mockInstaller) Exists(kind, namespace, name string) bool {
	for _, r := range m.Existing {
		if r.Kind == kind && r.Metadata.Namespace == namespace && r.Metadata.Name == name {
			return true
		}
	}
	return false
}

func (m *mockInstaller) Add(kind, namespace, name string) {
	m.Existing = append(m.Existing, glooinstall.ResourceType{
		TypeMeta: metav1.TypeMeta{
			Kind: kind,
		},
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	})
}

func (m *mockInstaller) WasDeleted(kind, namespace, name string) bool {
	for _, r := range m.Deleted {
		if r.Kind == kind && r.Metadata.Namespace == namespace && r.Metadata.Name == name {
			return true
		}
	}
	return false
}

func (m *mockInstaller) getResources(reader io.Reader) (res []glooinstall.ResourceType, err error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	for _, manifest := range strings.Split(string(bytes), "---") {
		if strings.Trim(manifest, " ") == "" {
			continue
		}
		var resource glooinstall.ResourceType
		if err := yaml.Unmarshal([]byte(manifest), &resource); err != nil {
			return nil, err
		}
		res = append(res, resource)
	}
	return
}

var mesh = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "my-mesh",
		Namespace: testNamespace,
	},
	MeshType: &v1.Mesh_AwsAppMesh{
		AwsAppMesh: &v1.AwsAppMesh{
			AwsSecret: &core.ResourceRef{
				Namespace: testNamespace,
				Name:      "aws-secret",
			},
			Region:           "us-east-1",
			EnableAutoInject: true,
			InjectionSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"inject": "true",
						},
					},
				},
			},
			VirtualNodeLabel: "vn",
		},
	},
}

var appMeshNoAutoInjection = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "my-mesh",
		Namespace: testNamespace,
	},
	MeshType: &v1.Mesh_AwsAppMesh{
		AwsAppMesh: &v1.AwsAppMesh{
			AwsSecret: &core.ResourceRef{
				Namespace: testNamespace,
				Name:      "aws-secret",
			},
			Region:           "us-east-1",
			EnableAutoInject: false,
		},
	},
}

var istio = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "my-istio-mesh",
		Namespace: testNamespace,
	},
	MeshType: &v1.Mesh_Istio{Istio: &v1.IstioMesh{}},
}
