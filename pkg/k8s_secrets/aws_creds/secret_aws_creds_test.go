package aws_creds_test

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/k8s_secrets/aws_creds"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SecretAwsCreds", func() {
	It("should convert AWS creds file to secret", func() {
		secretAwsCredsConverter := newSecretAwsCredsConverter()
		awsCredsMap := map[string]string{
			aws_creds.AWSAccessKeyID:     "foo",
			aws_creds.AWSSecretAccessKey: "bar",
		}
		expectedSecret := &k8s_core_types.Secret{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Labels:    map[string]string{aws_creds.AWSCredsSecretLabel: "true"},
				Name:      "secretName",
				Namespace: "secretNamespace",
			},
			Type:       k8s_core_types.SecretTypeOpaque,
			StringData: awsCredsMap,
		}
		secret, err := secretAwsCredsConverter.CredsFileToSecret(expectedSecret.GetName(), expectedSecret.GetNamespace(), "filename", "default")
		Expect(err).ToNot(HaveOccurred())
		Expect(secret).To(Equal(expectedSecret))
	})

	It("throw error when attempting to convert AWS creds file to secret", func() {
		secretAwsCredsConverter := newErrorSecretAwsCredsConverter()
		filename := "filename"
		profile := "default"
		_, err := secretAwsCredsConverter.CredsFileToSecret("name", "namespace", filename, profile)
		Expect(err).To(testutils.HaveInErrorChain(aws_creds.UnableToLoadAWSCreds(eris.New(""), filename, profile)))
	})

	It("should convert secret to AWS creds", func() {
		secretAwsCredsConverter := newSecretAwsCredsConverter()
		awsCredsMap := map[string][]byte{
			aws_creds.AWSAccessKeyID:     []byte("foo"),
			aws_creds.AWSSecretAccessKey: []byte("bar"),
		}
		secret := &k8s_core_types.Secret{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Labels:    map[string]string{aws_creds.AWSCredsSecretLabel: "true"},
				Name:      "secretName",
				Namespace: "secretNamespace",
			},
			Type: k8s_core_types.SecretTypeOpaque,
			Data: awsCredsMap,
		}
		creds, err := secretAwsCredsConverter.SecretToCreds(secret)
		Expect(err).ToNot(HaveOccurred())
		credsValue, err := creds.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(credsValue).To(Equal(credentials.Value{
			AccessKeyID:     "foo",
			SecretAccessKey: "bar",
		}))
	})

	It("should throw error when attempting to convert secret to AWS creds", func() {
		secretAwsCredsConverter := newSecretAwsCredsConverter()
		awsCredsMap := map[string][]byte{
			aws_creds.AWSAccessKeyID: []byte("foo"),
		}
		secret := &k8s_core_types.Secret{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Labels:    map[string]string{aws_creds.AWSCredsSecretLabel: "true"},
				Name:      "secretName",
				Namespace: "secretNamespace",
			},
			Type: k8s_core_types.SecretTypeOpaque,
			Data: awsCredsMap,
		}
		_, err := secretAwsCredsConverter.SecretToCreds(secret)
		Expect(err).To(testutils.HaveInErrorChain(aws_creds.MalformedAWSCredsSecret(aws_creds.AWSSecretAccessKey)))
	})
})

func newSecretAwsCredsConverter() aws_creds.SecretAwsCredsConverter {
	return aws_creds.NewSecretAwsCredsConverter(
		func(filename, profile string) *credentials.Credentials {
			return credentials.NewCredentials(awsCredsProviderMock{})
		},
	)
}

func newErrorSecretAwsCredsConverter() aws_creds.SecretAwsCredsConverter {
	return aws_creds.NewSecretAwsCredsConverter(
		func(filename, profile string) *credentials.Credentials {
			return credentials.NewCredentials(errAwsCredsProviderMock{})
		},
	)
}

type awsCredsProviderMock struct{}

func (t awsCredsProviderMock) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     "foo",
		SecretAccessKey: "bar",
	}, nil
}

func (t awsCredsProviderMock) IsExpired() bool {
	return false
}

type errAwsCredsProviderMock struct{}

func (t errAwsCredsProviderMock) Retrieve() (credentials.Value, error) {
	return credentials.Value{}, eris.New("")
}

func (t errAwsCredsProviderMock) IsExpired() bool {
	return false
}
