package generator

import (
	"github.com/solo-io/mesh-projects/pkg/project"
	"github.com/solo-io/mesh-projects/pkg/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest generation", func() {
	var (
		globalOptions = &project.GlobalOptions{
			BaseImageRepo:    version.BaseImageRepoName,
			BaseImageVersion: "some-version",
		}
	)

	var _ = Context("Standard simple app manifest", func() {
		It("should generate the expected manifest", func() {
			cfg := &ServiceManifestOutline{
				AppGroup: "app-group",
				AppName:  "app-name",
			}
			file, err := generateManifestFileContent(cfg, BasicServiceManifestTemplate)
			Expect(err).NotTo(HaveOccurred())
			fileContent, err := fileGenerationObjective.GetFileContent(file.Filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Content).To(Equal(fileContent))
			Expect(file.Filename).To(Equal("basic_service_manifest-app-name.yaml"))
		})
		It("should generate the expected go binary Makefile content", func() {
			cfg := &project.GoBinaryOutline{
				BinaryNameBase: "some-binary",
				ImageName:      "app-prefix-some-binary",
				BinaryDir:      "services/some-binary/cmd",
			}
			file, err := generateGoBinaryMakefileContent(cfg, GoBinaryMakefileBuildTemplate)
			Expect(err).NotTo(HaveOccurred())
			fileContent, err := fileGenerationObjective.GetFileContent(file.Filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Content).To(Equal(fileContent))
			Expect(file.Filename).To(Equal("go_binary_makefile_build.some-binary.partial.makefile"))
		})
		It("should generate the expected go binary Dockerfile", func() {
			cfg := &project.GoBinaryOutline{
				BinaryNameBase: "some-binary",
				ImageName:      "app-prefix-some-binary",
				BinaryDir:      "services/some-binary/cmd",
			}
			file, err := generateGoBinaryDockerfile(&GoBinaryOutlineTemplate{
				GoBinaryOutline: cfg,
				Global:          globalOptions,
			}, GoBinaryDockerfileTemplate)
			Expect(err).NotTo(HaveOccurred())
			fileContent, err := fileGenerationObjective.GetFileContent(file.Filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Content).To(Equal(fileContent))
			Expect(file.Filename).To(Equal("go_binary_dockerfile.some-binary.dockerfile"))
		})
		It("should generate the expected go binary Dockerfile with a base image", func() {
			cfg := &project.GoBinaryOutline{
				BinaryNameBase: "some-binary",
				ImageName:      "app-prefix-some-binary",
				BinaryDir:      "services/some-binary/cmd",
			}
			file, err := generateGoBinaryDockerfile(&GoBinaryOutlineTemplate{
				GoBinaryOutline: cfg,
				Global:          globalOptions,
			}, GoBinaryDockerfileWithCommonBaseImageTemplate)
			Expect(err).NotTo(HaveOccurred())
			fileContent, err := fileGenerationObjective.GetFileContent(file.Filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Content).To(Equal(fileContent))
			Expect(file.Filename).To(Equal("go_binary_dockerfile_with_base_image.some-binary.dockerfile"))
		})
	})
})
