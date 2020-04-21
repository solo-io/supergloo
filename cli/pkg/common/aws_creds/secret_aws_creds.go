package aws_creds

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rotisserie/eris"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TODO unify Secret type naming
	AWSSecretType = "solo.io/aws-credentials"
	// Uniquely identifies the AWS account
	AWSAccessKeyID     = "aws_access_key_id"
	AWSSecretAccessKey = "aws_secret_access_key"
)

var (
	UnableToLoadAWSCreds = func(err error, filename, profile string) error {
		return eris.Wrapf(err, "Unable to load AWS creds from file: %s, and profile: %s", filename, profile)
	}
	MalformedAWSCredsSecret = func(key string) error {
		return eris.Errorf("Malformed AWS credentials secret: missing key: %s", key)
	}
)

type secretAwsCredsConverter struct {
	awsCredentialsLoader AwsCredentialsLoader
}

func NewSecretAwsCredsConverter(awsCredentialsLoader AwsCredentialsLoader) SecretAwsCredsConverter {
	return &secretAwsCredsConverter{awsCredentialsLoader: awsCredentialsLoader}
}

func DefaultSecretAwsCredsConverter() SecretAwsCredsConverter {
	return &secretAwsCredsConverter{awsCredentialsLoader: credentials.NewSharedCredentials}
}

type AwsCredentialsLoader func(filename, profile string) *credentials.Credentials

// If "credsProfile" is empty, AWS SDK will default it to "default"
func (s *secretAwsCredsConverter) CredsFileToSecret(
	secretName,
	secretNamespace,
	credsFilename,
	credsProfile string,
) (*k8s_core_types.Secret, error) {
	creds, err := s.awsCredentialsLoader(credsFilename, credsProfile).Get()
	if err != nil {
		return nil, UnableToLoadAWSCreds(err, credsFilename, credsProfile)
	}
	// Persist AWS secrets as the pair of AWSAccessKeyID and AWSSecretAccessKey kv's
	secretData := make(map[string]string)
	secretData[AWSAccessKeyID] = creds.AccessKeyID
	secretData[AWSSecretAccessKey] = creds.SecretAccessKey
	return &k8s_core_types.Secret{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Type:       AWSSecretType,
		StringData: secretData,
	}, nil
}

func (s *secretAwsCredsConverter) SecretToCreds(secret *k8s_core_types.Secret) (*credentials.Credentials, error) {
	accessKeyID, ok := secret.Data[AWSAccessKeyID]
	if !ok {
		return nil, MalformedAWSCredsSecret(AWSAccessKeyID)
	}
	secretAccessKey, ok := secret.Data[AWSSecretAccessKey]
	if !ok {
		return nil, MalformedAWSCredsSecret(AWSSecretAccessKey)
	}

	return credentials.NewCredentials(newProvider(func() (credentials.Value, error) {
		return credentials.Value{
			AccessKeyID:     string(accessKeyID),
			SecretAccessKey: string(secretAccessKey),
		}, nil
	})), nil
}

func newProvider(retrieveFunc func() (credentials.Value, error)) *provider {
	return &provider{retrieveFunc: retrieveFunc}
}

type provider struct {
	retrieveFunc func() (credentials.Value, error)
}

func (p provider) Retrieve() (credentials.Value, error) {
	return p.retrieveFunc()
}

// User provided AWS credentials are assumed to be valid. If they are expired, the user must re-register the REST API with valid credentials.
func (p provider) IsExpired() bool {
	return false
}
