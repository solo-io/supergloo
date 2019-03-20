package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddAwsSecretFlags(set *pflag.FlagSet, secret *options.AwsSecret) {
	set.StringVar(&secret.AccessKeyId, "access-key-id", "", "AWS Access Key ID")
	set.StringVar(&secret.SecretAccessKey, "secret-access-key", "", "AWS Secret Access Key")
	set.StringVarP(&secret.CredentialsFileLocation, "file", "f", "", "path to the AWS credentials file")
	set.StringVar(&secret.CredentialsFileProfile, "profile", "", "name of the desired AWS profile in the credentials file")
}
