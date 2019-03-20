package surveyutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("Creating AWS Secrets", func() {

	Context("read aws credentials from file", func() {

		It("fills in the expected options", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("How do you want to provide the AWS credentials that will be stored in this secret?")
				c.PressDown()
				c.SendLine("") // select "Load from credentials file"
				c.ExpectString("Credential file location")
				c.SendLine("testfiles/aws_credentials_correct")
				c.ExpectEOF()
			}, func() {
				var opts options.Options
				err := surveyutils.SurveyAwsCredentials(&opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.AwsSecret.AccessKeyId).To(BeEquivalentTo("my-key"))
				Expect(opts.AwsSecret.SecretAccessKey).To(BeEquivalentTo("abcd1234"))
			})
		})

		It("fills in the expected options", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("How do you want to provide the AWS credentials that will be stored in this secret?")
				c.PressDown()
				c.SendLine("") // select "Load from credentials file"
				c.ExpectString("Credential file location")
				c.SendLine("testfiles/aws_credentials_correct_multiple")
				c.ExpectString("Found multiple profiles. Which one to use?")
				c.PressDown()
				c.SendLine("") // select profile 'another-profile
				c.ExpectEOF()
			}, func() {
				var opts options.Options
				err := surveyutils.SurveyAwsCredentials(&opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.AwsSecret.AccessKeyId).To(BeEquivalentTo("other-key"))
				Expect(opts.AwsSecret.SecretAccessKey).To(BeEquivalentTo("other9876"))
			})
		})

		It("fails on malformed credentials file", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("How do you want to provide the AWS credentials that will be stored in this secret?")
				c.PressDown()
				c.SendLine("") // select "Load from credentials file"
				c.ExpectString("Credential file location")
				c.SendLine("testfiles/aws_credentials_malformed")
				c.ExpectEOF()
			}, func() {
				var opts options.Options
				err := surveyutils.SurveyAwsCredentials(&opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find any profile section in credentials file"))
			})
		})

		It("fails on non-existing credentials file", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("How do you want to provide the AWS credentials that will be stored in this secret?")
				c.PressDown()
				c.SendLine("") // select "Load from credentials file"
				c.ExpectString("Credential file location")
				c.SendLine("non-existing-files/i-do-not-exist")
				c.ExpectEOF()
			}, func() {
				var opts options.Options
				err := surveyutils.SurveyAwsCredentials(&opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find credentials file: non-existing-files/i-do-not-exist"))
			})
		})
	})

	Context("read aws credentials from input", func() {

		It("fills in the expected options", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("How do you want to provide the AWS credentials that will be stored in this secret?")
				c.SendLine("") // select "Type them in"
				c.ExpectString("Access Key ID")
				c.SendLine("my-access-key")
				c.ExpectString("Secret Access Key")
				c.SendLine("my-secret")
				c.ExpectEOF()
			}, func() {
				var opts options.Options
				err := surveyutils.SurveyAwsCredentials(&opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.AwsSecret.AccessKeyId).To(BeEquivalentTo("my-access-key"))
				Expect(opts.AwsSecret.SecretAccessKey).To(BeEquivalentTo("my-secret"))
			})
		})
	})

})
