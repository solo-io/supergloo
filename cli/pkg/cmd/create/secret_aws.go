package create

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

func awsCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: `Create an AWS secret`,
		Long: `Creates a secret holding AWS access credentials. You can provide the access-key-id and secret-access-key 
either directly via the correspondent flags, or by passing the location of an AWS credentials file.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyAwsCredentials(opts); err != nil {
					return err
				}

				if err := surveyutils.SurveyMetadata("secret", &opts.Metadata); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			// Ensures either both credentials XOR a file was provided
			if err := validateFlags(opts); err != nil {
				return err
			}

			if opts.AwsSecret.CredentialsFileLocation != "" {
				if err := getCredentialsFromFile(opts); err != nil {
					return err
				}
			}

			// Check whether the given credentials can be used to read/write AWS App Mesh resources
			if err := verifyCredentials(opts); err != nil {
				return err
			}

			secret := &gloov1.Secret{
				Metadata: opts.Metadata,
				Kind: &gloov1.Secret_Aws{
					Aws: &gloov1.AwsSecret{
						AccessKey: opts.AwsSecret.AccessKeyId,
						SecretKey: opts.AwsSecret.SecretAccessKey,
					},
				},
			}

			secret, err := clients.MustSecretClient().Write(secret, skclients.WriteOpts{Ctx: opts.Ctx})
			if err != nil {
				return errors.Wrapf(err, "writing secret to storage")
			}

			helpers.PrintSecrets(gloov1.SecretList{secret}, opts.OutputType)

			return nil
		},
	}

	flagutils.AddAwsSecretFlags(cmd.PersistentFlags(), &opts.AwsSecret)

	return cmd
}

func validateFlags(opts *options.Options) error {
	if opts.Metadata.Name == "" {
		return errors.Errorf("name cannot be empty, provide with --name flag")
	}

	if opts.AwsSecret.AccessKeyId != "" && opts.AwsSecret.SecretAccessKey == "" ||
		opts.AwsSecret.AccessKeyId == "" && opts.AwsSecret.SecretAccessKey != "" ||
		opts.AwsSecret.AccessKeyId != "" && opts.AwsSecret.CredentialsFileLocation != "" ||
		opts.AwsSecret.AccessKeyId == "" && opts.AwsSecret.CredentialsFileLocation == "" {
		return errors.Errorf("you must provide either both the --access-key-id and --secret-access-key flags or " +
			"a credentials file via the -f flag")
	}
	return nil
}

// Parses the credentials file and sets the correspondent values in the given options
func getCredentialsFromFile(opts *options.Options) error {
	sections, err := helpers.ParseAwsCredentialsFile(opts.AwsSecret.CredentialsFileLocation)
	if err != nil {
		return err
	}

	// The aws credentials file is a .ini config file with one section per profile, we have to select one
	var section *ini.Section
	switch len(sections) {
	case 0:
		return errors.Errorf("could not find any profile section in credentials file")
	case 1:
		section = sections[0]
		if profile := opts.AwsSecret.CredentialsFileProfile; profile != "" && section.Name() != profile {
			return errors.Errorf("could not find profile [%s] in credentials file", profile)
		}
	default:
		profile := opts.AwsSecret.CredentialsFileProfile
		if profile == "" {
			return errors.Errorf("found multiple profiles in credentials file. Please select the one " +
				"you would like to use via the --profile flag")
		}

		for _, s := range sections {
			if s.Name() == profile {
				section = s
			}
		}

		if section == nil {
			return errors.Errorf("could not find profile [%s] in credentials file", profile)
		}
	}

	if err := helpers.SetAwsCredentialsFromSection(opts, section); err != nil {
		return err
	}
	return nil
}

func verifyCredentials(opts *options.Options) error {
	appmeshClient, err := clients.NewAppmeshClient(opts.AwsSecret.AccessKeyId, opts.AwsSecret.SecretAccessKey, "us-east-1")
	if err != nil {
		return errors.Wrapf(err, "failed to create aws session with provided credentials")
	}

	// Check if we can connect to appmesh with the provided credentials
	if _, err := appmeshClient.ListMeshes(nil); err != nil {
		return errors.Wrapf(err, "unable to access the AWS App Mesh service using the provided credentials. "+
			"Make sure they are associated with an IAM user that has permissions to access and modify AWS App "+
			"Mesh resources. Underlying error is")
	}
	return nil
}
