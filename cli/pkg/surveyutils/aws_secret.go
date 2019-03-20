package surveyutils

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"gopkg.in/ini.v1"

	"github.com/solo-io/gloo/pkg/cliutil"
)

// Survey to get AWS access credentials either from stdin or from a credential file.
func SurveyAwsCredentials(opts *options.Options) error {
	readFromInput, loadCredFile := "Type them in", "Load from credentials file"
	var choice string
	if err := cliutil.ChooseFromList(
		"How do you want to provide the AWS credentials that will be stored in this secret?",
		&choice,
		[]string{readFromInput, loadCredFile},
	); err != nil {
		return err
	}

	if strings.Compare(choice, readFromInput) == 0 {
		if err := cliutil.GetStringInput("Access Key ID", &opts.AwsSecret.AccessKeyId); err != nil {
			return err
		}
		if err := cliutil.GetStringInput("Secret Access Key", &opts.AwsSecret.SecretAccessKey); err != nil {
			return err
		}
	} else {
		credentialFileLocation, err := promptForCredentialFileLocation()
		if err != nil {
			return err
		}

		sections, err := helpers.ParseAwsCredentialsFile(credentialFileLocation)
		if err != nil {
			return err
		}

		// The aws credentials file is a .ini config file with one section per profile, we have to select one
		var section *ini.Section
		switch len(sections) {
		case 0:
			return errors.New("could not find any profile section in credentials file")
		case 1:
			section = sections[0]
			break
		default:
			section, err = promptForProfileSelection(sections)
		}

		if err := helpers.SetAwsCredentialsFromSection(opts, section); err != nil {
			return err
		}
	}
	return nil
}

func promptForCredentialFileLocation() (string, error) {
	defaultConfigFileLocation := "~/.aws/credentials"

	// Check if custom location is set via env
	if envLocation := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); envLocation != "" {
		defaultConfigFileLocation = envLocation
	}

	var configFileLocation string
	if err := cliutil.GetStringInput(
		fmt.Sprintf("Credential file location [default: %s]", defaultConfigFileLocation),
		&configFileLocation,
	); err != nil {
		return "", err
	}
	if configFileLocation == "" {
		// expands '~' on any OS
		expanded, err := homedir.Expand(defaultConfigFileLocation)
		if err != nil {
			return "", err
		}
		configFileLocation = expanded
	}

	return configFileLocation, nil
}

func promptForProfileSelection(sections []*ini.Section) (*ini.Section, error) {
	var profiles []string
	profileMap := make(map[string]*ini.Section)
	for _, s := range sections {
		profiles = append(profiles, s.Name())
		profileMap[s.Name()] = s
	}

	var profile string
	if err := cliutil.ChooseFromList(
		"Found multiple profiles. Which one to use?", &profile, profiles); err != nil {
		return nil, err
	}

	return profileMap[profile], nil
}
