package create_test

import (
	"fmt"
	"io/ioutil"
	"os"

	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers/mocks"
	"github.com/solo-io/supergloo/cli/test/utils"
)

var _ = Describe("Create Secret Aws CLI Command", func() {

	const CredentialValidationFailedMessage = "you must provide either both the --access-key-id and " +
		"--secret-access-key flags or a credentials file via the -f flag"

	var (
		successMock, failMock *mocks.MockAppmesh
		ctrl                  *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)

		successMock = mocks.NewMockAppmesh(ctrl)
		successMock.EXPECT().ListMeshes(nil).Return(&appmesh.ListMeshesOutput{}, nil).AnyTimes()

		failMock = mocks.NewMockAppmesh(ctrl)
		failMock.EXPECT().ListMeshes(nil).Return(nil, errors.Errorf("mock returns error")).AnyTimes()

		clients.UseMemoryClients()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("flag validation works as expected", func() {

		It("fails if no name was provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --access-key-id %s --secret-access-key %s",
				"key-1",
				"asdf",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name cannot be empty, provide with --name flag"))
		})

		It("fails if no credential flags were provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --name %s", "my-secret",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(CredentialValidationFailedMessage))
		})

		It("fails if no secret access key was provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --name %s --access-key-id %s", "my-secret", "key-1",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(CredentialValidationFailedMessage))
		})

		It("fails if no access key ID was provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --name %s --secret-access-key %s", "my-secret", "asdf2345",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(CredentialValidationFailedMessage))
		})

		It("fails if both credentials flags and file flag were provided", func() {
			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --name %s --access-key-id %s --secret-access-key %s -f %s",
				"my-secret", "key-1", "asdf2345", "dir/some-file",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(CredentialValidationFailedMessage))
		})
	})

	Describe("create a secret from credentials flags", func() {

		It("succeeds with valid credentials", func() {
			clients.UseAppmeshMock(successMock)

			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --name %s --access-key-id %s --secret-access-key %s",
				"my-secret", "valid-key", "valid-secret",
			))
			Expect(err).NotTo(HaveOccurred())

			secret, err := clients.MustSecretClient().Read("supergloo-system", "my-secret", skclients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Metadata.Name).To(BeEquivalentTo("my-secret"))
			Expect(secret.Metadata.Namespace).To(BeEquivalentTo("supergloo-system"))

			awsSecret, ok := secret.Kind.(*v1.Secret_Aws)
			Expect(ok).To(BeTrue())
			Expect(awsSecret.Aws.AccessKey).To(BeEquivalentTo("valid-key"))
			Expect(awsSecret.Aws.SecretKey).To(BeEquivalentTo("valid-secret"))
		})

		It("fails with invalid credentials", func() {
			clients.UseAppmeshMock(failMock)

			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --name %s --access-key-id %s --secret-access-key %s",
				"my-secret", "invalid-key", "invalid-secret",
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to access the AWS App Mesh service"))
		})
	})

	Describe("create a secret from credentials file", func() {

		var credentialsFile *os.File

		Context("valid credentials file with one profile", func() {

			BeforeEach(func() {
				clients.UseAppmeshMock(successMock)
				credentialsFile = createFile(oneProfileValid)
			})

			AfterEach(func() {
				_ = os.Remove(credentialsFile.Name())
			})

			It("succeeds when not providing any --profile flag", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"create secret aws --name %s --namespace %s --file %s",
					"my-secret", "my-ns", credentialsFile.Name(),
				))
				Expect(err).NotTo(HaveOccurred())

				secret, err := clients.MustSecretClient().Read("my-ns", "my-secret", skclients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(secret).NotTo(BeNil())
				Expect(secret.Metadata.Name).To(BeEquivalentTo("my-secret"))
				Expect(secret.Metadata.Namespace).To(BeEquivalentTo("my-ns"))

				awsSecret, ok := secret.Kind.(*v1.Secret_Aws)
				Expect(ok).To(BeTrue())
				Expect(awsSecret.Aws.AccessKey).To(BeEquivalentTo("ABCD2134"))
				Expect(awsSecret.Aws.SecretKey).To(BeEquivalentTo("cmeLK8Wc4pgSVQsfnNJWc4pgSVQd8"))
			})

			It("fails when providing a non-existing profile", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"create secret aws --name %s --file %s --profile %s",
					"my-secret", credentialsFile.Name(), "non-existing",
				))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find profile [non-existing] in credentials file"))
			})
		})

		Context("valid credentials file with multiple profiles", func() {

			BeforeEach(func() {
				clients.UseAppmeshMock(successMock)
				credentialsFile = createFile(twoProfilesValid)
			})

			AfterEach(func() {
				_ = os.Remove(credentialsFile.Name())
			})

			It("succeeds when profile flag is provided", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"create secret aws --name %s --file %s --profile %s",
					"my-secret", credentialsFile.Name(), "another",
				))
				Expect(err).NotTo(HaveOccurred())

				secret, err := clients.MustSecretClient().Read("supergloo-system", "my-secret", skclients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(secret).NotTo(BeNil())
				Expect(secret.Metadata.Name).To(BeEquivalentTo("my-secret"))
				Expect(secret.Metadata.Namespace).To(BeEquivalentTo("supergloo-system"))

				awsSecret, ok := secret.Kind.(*v1.Secret_Aws)
				Expect(ok).To(BeTrue())
				Expect(awsSecret.Aws.AccessKey).To(BeEquivalentTo("ZYXW9876"))
				Expect(awsSecret.Aws.SecretKey).To(BeEquivalentTo("srGfWc4pgSVQ8WWc4pgSVQk8BODMx"))
			})

			It("fails when no profile flag is provided", func() {

				err := utils.Supergloo(fmt.Sprintf(
					"create secret aws --name %s --file %s",
					"my-secret", credentialsFile.Name(),
				))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("found multiple profiles in credentials file. " +
					"Please select the one you would like to use via the --profile flag"))
			})

			It("fails when specifying a non-existing profile", func() {
				err := utils.Supergloo(fmt.Sprintf(
					"create secret aws --name %s --file %s --profile %s",
					"my-secret", credentialsFile.Name(), "non-existing",
				))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find profile [non-existing] in credentials file"))
			})
		})

		It("fails when the credentials file is invalid", func() {
			clients.UseAppmeshMock(successMock)
			credentialsFile = createFile(invalid)
			//noinspection GoUnhandledErrorResult
			defer os.Remove(credentialsFile.Name())

			err := utils.Supergloo(fmt.Sprintf(
				"create secret aws --name %s --file %s",
				"my-secret", credentialsFile.Name(),
			))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find any profile section in credentials file"))
		})
	})
})

// Creates a temporary AWS credentials file
func createFile(content string) *os.File {
	credentialsFile, err := ioutil.TempFile("", "aws_creds")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	_, err = credentialsFile.Write([]byte(content))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return credentialsFile
}

var oneProfileValid = `
[default]
aws_access_key_id = ABCD2134
aws_secret_access_key = cmeLK8Wc4pgSVQsfnNJWc4pgSVQd8
`

var twoProfilesValid = `
[default]
aws_access_key_id = ABCD2134
aws_secret_access_key = cmeLK8Wc4pgSVQsfnNJWc4pgSVQd8
[another]
aws_access_key_id = ZYXW9876
aws_secret_access_key = srGfWc4pgSVQ8WWc4pgSVQk8BODMx
`

var invalid = `
aws_access_key_id = ABCD2134
aws_secret_access_key = cmeLK8Wc4pgSVQsfnNJWc4pgSVQd8
`
