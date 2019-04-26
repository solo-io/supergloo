package surveyutils_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("RegisterAppmesh", func() {

	BeforeEach(func() {
		clients.UseMemoryClients()
	})

	Describe("successful scenarios", func() {

		BeforeEach(func() {
			_, err := clients.MustSecretClient().Write(&gloov1.Secret{
				Metadata: core.Metadata{
					Namespace: "default",
					Name:      "aws",
				},
				Kind: &gloov1.Secret_Aws{
					Aws: &gloov1.AwsSecret{
						AccessKey: "abc123",
						SecretKey: "def456",
					},
				},
			}, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			_, err = clients.MustKubeClient().CoreV1().ConfigMaps("default").Create(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "custom-patch"},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("fills in the expected options when auto-injection is disabled", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("Choose a secret that contains the AWS credentials to " +
					"connect to AWS App Mesh (hint: you can easily create one with 'supergloo create secret aws -i')")
				c.SendLine("")
				c.ExpectString("In which AWS region do you want AWS App Mesh resources to be created?")
				c.PressDown()
				c.SendLine("") // select "ap-northeast-2"
				c.ExpectString("Do you want SuperGloo to auto-inject your pods with the AWS App Mesh sidecar proxies?")
				c.PressDown()
				c.SendLine("") // select "no"
				c.ExpectEOF()
			}, func() {
				var opts options.Options
				err := surveyutils.SurveyAppmeshRegistration(context.Background(), &opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.RegisterAppMesh.Secret.Name).To(BeEquivalentTo("aws"))
				Expect(opts.RegisterAppMesh.Secret.Namespace).To(BeEquivalentTo("default"))
				Expect(opts.RegisterAppMesh.Region).To(BeEquivalentTo("ap-northeast-2"))
				Expect(opts.RegisterAppMesh.EnableAutoInjection).To(BeEquivalentTo("false"))
			})
		})

		Context("auto-injection is enabled", func() {

			It("fills in the expected options when using namespace selector", func() {
				testutil.ExpectInteractive(func(c *testutil.Console) {
					c.ExpectString("Choose a secret that contains the AWS credentials to " +
						"connect to AWS App Mesh (hint: you can easily create one with 'supergloo create secret aws -i')")
					c.SendLine("")
					c.ExpectString("In which AWS region do you want AWS App Mesh resources to be created?")
					c.PressDown()
					c.PressDown()
					c.SendLine("") // select "ap-southeast-1"
					c.ExpectString("Do you want SuperGloo to auto-inject your pods with the AWS App Mesh sidecar proxies?")
					c.SendLine("") // select "yes"
					c.ExpectString("How you want SuperGloo to select which pods will be auto-injected?")
					c.SendLine("") // select namespace selector
					c.ExpectString("add a namespace (choose <done> to finish): ")
					c.PressDown()
					c.SendLine("") // select a namespace
					c.ExpectString("add a namespace (choose <done> to finish): ")
					c.SendLine("") // end namespace selection
					c.ExpectString("SuperGloo looks for the patch that will be applied to the pods matching the selection criteria in a config map. " +
						"Do you want to use the default one provided by supergloo or provide your own?")
					c.SendLine("") // use default
					c.ExpectString("Select the key of the pod label that SuperGloo will use to assign your pods to a VirtualNode")
					c.SendLine("vn")
					c.ExpectEOF()
				}, func() {
					var opts options.Options
					err := surveyutils.SurveyAppmeshRegistration(context.Background(), &opts)
					Expect(err).NotTo(HaveOccurred())
					Expect(opts.RegisterAppMesh.Secret.Name).To(BeEquivalentTo("aws"))
					Expect(opts.RegisterAppMesh.Secret.Namespace).To(BeEquivalentTo("default"))
					Expect(opts.RegisterAppMesh.Region).To(BeEquivalentTo("ap-southeast-1"))
					Expect(opts.RegisterAppMesh.EnableAutoInjection).To(BeEquivalentTo("true"))
					Expect(opts.RegisterAppMesh.PodSelector.SelectedNamespaces).To(ConsistOf("default"))
					Expect(opts.RegisterAppMesh.VirtualNodeLabel).To(BeEquivalentTo("vn"))
					Expect(opts.RegisterAppMesh.ConfigMap.Name).To(BeEmpty())
					Expect(opts.RegisterAppMesh.ConfigMap.Namespace).To(BeEmpty())
				})
			})

			It("fills in the expected options when using label selector", func() {
				testutil.ExpectInteractive(func(c *testutil.Console) {
					c.ExpectString("Choose a secret that contains the AWS credentials to " +
						"connect to AWS App Mesh (hint: you can easily create one with 'supergloo create secret aws -i')")
					c.SendLine("")
					c.ExpectString("In which AWS region do you want AWS App Mesh resources to be created?")
					c.PressDown()
					c.PressDown()
					c.SendLine("") // select "ap-southeast-1"
					c.ExpectString("Do you want SuperGloo to auto-inject your pods with the AWS App Mesh sidecar proxies?")
					c.SendLine("") // select "yes"
					c.ExpectString("How you want SuperGloo to select which pods will be auto-injected?")
					c.PressDown()
					c.SendLine("") // select label selector
					c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish:")
					c.SendLine("key=value")
					c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish:")
					c.SendLine("") // end label selection
					c.ExpectString("SuperGloo looks for the patch that will be applied to the pods matching the selection criteria in a config map. " +
						"Do you want to use the default one provided by supergloo or provide your own?")
					c.SendLine("") // use default
					c.ExpectString("Select the key of the pod label that SuperGloo will use to assign your pods to a VirtualNode")
					c.SendLine("vn")
					c.ExpectEOF()
				}, func() {
					var opts options.Options
					err := surveyutils.SurveyAppmeshRegistration(context.Background(), &opts)
					Expect(err).NotTo(HaveOccurred())
					Expect(opts.RegisterAppMesh.Secret.Name).To(BeEquivalentTo("aws"))
					Expect(opts.RegisterAppMesh.Secret.Namespace).To(BeEquivalentTo("default"))
					Expect(opts.RegisterAppMesh.Region).To(BeEquivalentTo("ap-southeast-1"))
					Expect(opts.RegisterAppMesh.EnableAutoInjection).To(BeEquivalentTo("true"))
					Expect(opts.RegisterAppMesh.PodSelector.SelectedLabels).To(BeEquivalentTo(map[string]string{"key": "value"}))
					Expect(opts.RegisterAppMesh.VirtualNodeLabel).To(BeEquivalentTo("vn"))
					Expect(opts.RegisterAppMesh.ConfigMap.Name).To(BeEmpty())
					Expect(opts.RegisterAppMesh.ConfigMap.Namespace).To(BeEmpty())
				})
			})

			It("fills in the expected options when using a custom config map", func() {
				testutil.ExpectInteractive(func(c *testutil.Console) {
					c.ExpectString("Choose a secret that contains the AWS credentials to " +
						"connect to AWS App Mesh (hint: you can easily create one with 'supergloo create secret aws -i')")
					c.SendLine("")
					c.ExpectString("In which AWS region do you want AWS App Mesh resources to be created?")
					c.PressDown()
					c.PressDown()
					c.SendLine("") // select "ap-southeast-1"
					c.ExpectString("Do you want SuperGloo to auto-inject your pods with the AWS App Mesh sidecar proxies?")
					c.SendLine("") // select "yes"
					c.ExpectString("How you want SuperGloo to select which pods will be auto-injected?")
					c.PressDown()
					c.SendLine("") // select label selector
					c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish:")
					c.SendLine("key=value")
					c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish:")
					c.SendLine("") // end label selection
					c.ExpectString("SuperGloo looks for the patch that will be applied to the pods matching the selection criteria in a config map. " +
						"Do you want to use the default one provided by supergloo or provide your own?")
					c.PressDown()
					c.SendLine("") // use custom
					c.ExpectString("Select the config map that you would like to use")
					c.SendLine("") // select default.custom-patch
					c.ExpectString("Select the key of the pod label that SuperGloo will use to assign your pods to a VirtualNode")
					c.SendLine("vn")
					c.ExpectEOF()
				}, func() {
					var opts options.Options
					err := surveyutils.SurveyAppmeshRegistration(context.Background(), &opts)
					Expect(err).NotTo(HaveOccurred())
					Expect(opts.RegisterAppMesh.Secret.Name).To(BeEquivalentTo("aws"))
					Expect(opts.RegisterAppMesh.Secret.Namespace).To(BeEquivalentTo("default"))
					Expect(opts.RegisterAppMesh.Region).To(BeEquivalentTo("ap-southeast-1"))
					Expect(opts.RegisterAppMesh.EnableAutoInjection).To(BeEquivalentTo("true"))
					Expect(opts.RegisterAppMesh.PodSelector.SelectedLabels).To(BeEquivalentTo(map[string]string{"key": "value"}))
					Expect(opts.RegisterAppMesh.VirtualNodeLabel).To(BeEquivalentTo("vn"))
					Expect(opts.RegisterAppMesh.ConfigMap.Name).To(BeEquivalentTo("custom-patch"))
					Expect(opts.RegisterAppMesh.ConfigMap.Namespace).To(BeEquivalentTo("default"))
				})
			})
		})
	})

	Describe("failure scenarios", func() {

		It("fails if no AWS secret could be found in the cluster", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectEOF()
			}, func() {
				var opts options.Options
				err := surveyutils.SurveyAppmeshRegistration(context.Background(), &opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find any AWS secret"))

			})
		})
	})
})
