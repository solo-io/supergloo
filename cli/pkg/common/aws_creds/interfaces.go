package aws_creds

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	k8s_core_types "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type SecretAwsCredsConverter interface {
	/*
		Convert an AWS credentials file (typically stored in ~/.aws/config) to an SMH formatted k8s Secret
		The secretName identifies the AWS account for which the credentials grant access.
	*/
	CredsFileToSecret(
		secretName,
		secretNamespace,
		credsFilename,
		credsProfile string,
	) (*k8s_core_types.Secret, error)

	// Convert an SMH secret to AWS credentials
	SecretToCreds(secret *k8s_core_types.Secret) (*credentials.Credentials, error)
}
