package register_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
)

var _ = Describe("AppMesh", func() {

	BeforeEach(func() {
		clients.UseMemoryClients()
	})

	Describe("flag validation works as expected", func() {

		It("fails if no name was provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"register appmesh --region %s --secret %s --auto-inject %s",
				"us-east-1",
				"DEFAULT.SECRET",
				"false",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name cannot be empty, provide with --name flag"))
		})

		It("fails if no region was provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"register appmesh --name %s --secret %s --auto-inject %s",
				"mesh-1",
				"DEFAULT.SECRET",
				"false",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("AWS region cannot be empty, provide with --region flag"))
		})

		It("fails if invalid region was provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"register appmesh --name %s --region %s --secret %s --auto-inject %s",
				"mesh-1",
				"us-somewhere-1",
				"DEFAULT.SECRET",
				"false",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid region. AWS App Mesh is currently available in the following regions"))
		})

		It("fails if no secret was provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"register appmesh --name %s --region %s --auto-inject %s",
				"mesh-1",
				"us-east-1",
				"false",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("you must provide a fully qualified secret name"))
		})

		It("fails if secret format is incorrect", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"register appmesh --name %s --region %s --secret %s --auto-inject %s",
				"mesh-1",
				"us-east-1",
				"just-secret-name.",
				"false",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("refs must be specified in the format <NAMESPACE>.<NAME>"))
		})

		It("fails if auto-inject flag value is incorrect", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"register appmesh --name %s --region %s --secret %s --auto-inject %s",
				"mesh-1",
				"us-east-1",
				"ns.name",
				"yes",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid value for --auto-inject flag. Must be either true or false"))
		})

		Context("auto-injection is enabled", func() {

			It("fails if no pod selector is specified", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"register appmesh --name %s --region %s --secret %s --virtual-node-label %s",
					"mesh-1",
					"us-east-1",
					"ns.secret-name",
					"vn",
				))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("you must provide a pod selector if auto-injection is enabled"))
			})

			It("correctly parses label selector", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"register appmesh --name %s --region %s --secret %s --select-labels %s --virtual-node-label %s",
					"mesh-1",
					"us-east-1",
					"ns.secret-name",
					"key=value",
					"vn",
				))
				Expect(err).NotTo(HaveOccurred())
			})

			It("correctly parses namespace selector", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"register appmesh --name %s --region %s --secret %s --select-namespaces %s --virtual-node-label %s",
					"mesh-1",
					"us-east-1",
					"ns.secret-name",
					"ns-1,ns-2",
					"vn",
				))
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails if no virtual node label is specified", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"register appmesh --name %s --region %s --secret %s --select-namespaces %s",
					"mesh-1",
					"us-east-1",
					"ns.secret-name",
					"ns-1,ns-2",
				))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("you must provide a virtual node label if auto-injection is enabled"))
			})

		})
	})

	It("correctly creates a mesh without auto-injection", func() {
		err := utils.Supergloo(fmt.Sprintf(
			"register appmesh --name %s --region %s --secret %s --auto-inject %s",
			"mesh-1",
			"us-east-1",
			"ns.secret-name",
			"false",
		))
		Expect(err).NotTo(HaveOccurred())

		mesh, err := clients.MustMeshClient().Read("supergloo-system", "mesh-1", skclients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(mesh.Metadata.Namespace).To(BeEquivalentTo("supergloo-system"))
		Expect(mesh.Metadata.Name).To(BeEquivalentTo("mesh-1"))

		appMesh := mesh.GetAwsAppMesh()
		Expect(appMesh).NotTo(BeNil())
		Expect(appMesh.Region).To(BeEquivalentTo("us-east-1"))
		Expect(appMesh.AwsSecret.Namespace).To(BeEquivalentTo("ns"))
		Expect(appMesh.AwsSecret.Name).To(BeEquivalentTo("secret-name"))
		Expect(appMesh.EnableAutoInject).To(BeFalse())
	})

	It("correctly creates a mesh with auto-injection (namespace selector)", func() {
		err := utils.Supergloo(fmt.Sprintf(
			"register appmesh --name %s --region %s --secret %s --select-namespaces %s --virtual-node-label %s",
			"mesh-1",
			"us-east-1",
			"ns.secret-name",
			"default,ns-1",
			"vn",
		))
		Expect(err).NotTo(HaveOccurred())

		mesh, err := clients.MustMeshClient().Read("supergloo-system", "mesh-1", skclients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(mesh.Metadata.Namespace).To(BeEquivalentTo("supergloo-system"))
		Expect(mesh.Metadata.Name).To(BeEquivalentTo("mesh-1"))

		appMesh := mesh.GetAwsAppMesh()
		Expect(appMesh).NotTo(BeNil())
		Expect(appMesh.Region).To(BeEquivalentTo("us-east-1"))
		Expect(appMesh.AwsSecret.Namespace).To(BeEquivalentTo("ns"))
		Expect(appMesh.AwsSecret.Name).To(BeEquivalentTo("secret-name"))
		Expect(appMesh.EnableAutoInject).To(BeTrue())
		Expect(appMesh.VirtualNodeLabel).To(BeEquivalentTo("vn"))

		nsSelector := appMesh.InjectionSelector.GetNamespaceSelector()
		Expect(nsSelector).NotTo(BeNil())
		Expect(nsSelector.Namespaces).To(BeEquivalentTo([]string{"default", "ns-1"}))
	})

	It("correctly creates a mesh with auto-injection (label selector)", func() {
		err := utils.Supergloo(fmt.Sprintf(
			"register appmesh --name %s --region %s --secret %s --select-labels %s --virtual-node-label %s",
			"mesh-1",
			"us-east-1",
			"ns.secret-name",
			"key=value",
			"vn",
		))
		Expect(err).NotTo(HaveOccurred())

		mesh, err := clients.MustMeshClient().Read("supergloo-system", "mesh-1", skclients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(mesh.Metadata.Namespace).To(BeEquivalentTo("supergloo-system"))
		Expect(mesh.Metadata.Name).To(BeEquivalentTo("mesh-1"))

		appMesh := mesh.GetAwsAppMesh()
		Expect(appMesh).NotTo(BeNil())
		Expect(appMesh.Region).To(BeEquivalentTo("us-east-1"))
		Expect(appMesh.AwsSecret.Namespace).To(BeEquivalentTo("ns"))
		Expect(appMesh.AwsSecret.Name).To(BeEquivalentTo("secret-name"))
		Expect(appMesh.EnableAutoInject).To(BeTrue())
		Expect(appMesh.VirtualNodeLabel).To(BeEquivalentTo("vn"))

		labelSelector := appMesh.InjectionSelector.GetLabelSelector()
		Expect(labelSelector).NotTo(BeNil())
		Expect(labelSelector.LabelsToMatch).To(BeEquivalentTo(map[string]string{"key": "value"}))
	})
})
