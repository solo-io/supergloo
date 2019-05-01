package appmesh

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	gloomocks "github.com/solo-io/gloo/projects/gloo/pkg/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	climocks "github.com/solo-io/supergloo/cli/pkg/helpers/mocks"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	. "github.com/onsi/gomega"
)

var _ = Describe("Validate App Mesh meshes", func() {

	var (
		kube         kubernetes.Interface
		secretClient *gloomocks.MockSecretClient
		ctrl         *gomock.Controller
		validator    Validator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)

		successMock := climocks.NewMockAppmesh(ctrl)
		successMock.EXPECT().ListMeshes(nil).Return(&appmesh.ListMeshesOutput{}, nil).AnyTimes()
		clients.UseAppmeshMock(successMock)

		secretClient = gloomocks.NewMockSecretClient(ctrl)
		secretClient.EXPECT().Read(testNamespace, "aws-secret", gomock.Any()).Return(awsSecret, nil).AnyTimes()
		secretClient.EXPECT().Read(testNamespace, gomock.Not("aws-secret"), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()

		kube = fake.NewSimpleClientset()
		_, err := kube.CoreV1().ConfigMaps(testNamespace).Create(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace, Name: webhookName}})
		Expect(err).NotTo(HaveOccurred())

		validator = NewAppMeshValidator(kube, secretClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("fails if mesh does not specify a region", func() {
		err := validator.Validate(context.TODO(), &meshNoRegion)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("region is required for AWS App Mesh"))
	})

	It("fails if mesh region is invalid", func() {
		err := validator.Validate(context.TODO(), &meshInvalidRegion)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid AWS region"))
	})

	It("fails if mesh does not specify an AWS secret", func() {
		err := validator.Validate(context.TODO(), &meshNoSecret)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("an AWS secret is required"))
	})

	It("fails if secret AWS secret does not exist", func() {
		err := validator.Validate(context.TODO(), &meshNonExistingSecret)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to retrieve secret"))
	})

	It("fails if we cannot connect to AWS App Mesh with the given secret", func() {
		failMock := climocks.NewMockAppmesh(ctrl)
		failMock.EXPECT().ListMeshes(nil).Return(nil, errors.New("error")).AnyTimes()
		clients.UseAppmeshMock(failMock)

		err := validator.Validate(context.TODO(), &autoInjectionMesh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to verify AWS credentials stored in secret"))
	})

	Context("auto-injection is enabled", func() {

		It("fails if mesh does not specify a pod selector", func() {
			err := validator.Validate(context.TODO(), &autoInjectionMeshNoSelector)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("InjectionSelector is required when EnableAutoInject==true"))
		})

		It("fails if mesh pod selector is of unsupported type", func() {
			err := validator.Validate(context.TODO(), &autoInjectionMeshUnsupportedSelector)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("upstream injection selectors are currently not supported"))
		})

		It("fails if the mesh specifies a custom SidecarPatchConfigMap that does not exist", func() {
			err := validator.Validate(context.TODO(), &autoInjectionMeshCustomConfigMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to find SidecarPatchConfigMap"))
		})

		It("fails if mesh does not specify a VirtualNodeLabel", func() {
			err := validator.Validate(context.TODO(), &autoInjectionMeshNoVirtualNodeLabel)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("VirtualNodeLabel is required when EnableAutoInject==true"))
		})

		It("fails if VirtualNodeLabel has an invalid format", func() {
			err := validator.Validate(context.TODO(), &autoInjectionMeshInvalidVirtualNodeLabel)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid VirtualNodeLabel format"))
		})
	})

})

var awsSecret = &gloov1.Secret{
	Metadata: core.Metadata{
		Namespace: testNamespace,
		Name:      "aws-secret",
	},
	Kind: &gloov1.Secret_Aws{
		Aws: &gloov1.AwsSecret{
			AccessKey: "abcd2134",
			SecretKey: "fdea546",
		},
	},
}

var meshNoRegion = v1.AwsAppMesh{
	AwsSecret: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      "aws-secret",
	},
	EnableAutoInject: false,
}
var meshInvalidRegion = v1.AwsAppMesh{
	AwsSecret: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      "aws-secret",
	},
	Region:           "us-north-9",
	EnableAutoInject: false,
}

var meshNoSecret = v1.AwsAppMesh{
	Region:           "us-east-1",
	EnableAutoInject: false,
}
var meshNonExistingSecret = v1.AwsAppMesh{
	AwsSecret: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      "i-do-not-exist",
	},
	Region:           "us-east-1",
	EnableAutoInject: false,
}

var autoInjectionMesh = v1.AwsAppMesh{
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
	SidecarPatchConfigMap: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      webhookName,
	},
}

var autoInjectionMeshNoSelector = v1.AwsAppMesh{
	AwsSecret: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      "aws-secret",
	},
	Region:           "us-east-1",
	EnableAutoInject: true,
	VirtualNodeLabel: "vn",
	SidecarPatchConfigMap: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      webhookName,
	},
}

var autoInjectionMeshUnsupportedSelector = v1.AwsAppMesh{
	AwsSecret: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      "aws-secret",
	},
	Region:           "us-east-1",
	EnableAutoInject: true,
	InjectionSelector: &v1.PodSelector{
		SelectorType: &v1.PodSelector_UpstreamSelector_{
			UpstreamSelector: &v1.PodSelector_UpstreamSelector{},
		},
	},
	VirtualNodeLabel: "vn",
	SidecarPatchConfigMap: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      webhookName,
	},
}

var autoInjectionMeshCustomConfigMap = v1.AwsAppMesh{
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
	SidecarPatchConfigMap: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      "my-wn-config-map",
	},
}

var autoInjectionMeshNoVirtualNodeLabel = v1.AwsAppMesh{
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
	SidecarPatchConfigMap: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      webhookName,
	},
}

var autoInjectionMeshInvalidVirtualNodeLabel = v1.AwsAppMesh{
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
	VirtualNodeLabel: "inv@lid",
	SidecarPatchConfigMap: &core.ResourceRef{
		Namespace: testNamespace,
		Name:      webhookName,
	},
}
