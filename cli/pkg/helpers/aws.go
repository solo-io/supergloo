package helpers

import (
	"os"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"gopkg.in/ini.v1"
)

func ParseAwsCredentialsFile(path string) ([]*ini.Section, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("could not find credentials file: %s", path)
		}
		return nil, errors.Wrapf(err, "unexpected error while loading credentials file: "+path)
	}

	credentialFile, err := ini.Load(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse credentials file %s", path)
	}

	// Remove DEFAULT section (this is the top level section, not the 'default' profile)
	sections := credentialFile.Sections()
	for i, s := range sections {
		if s.Name() == ini.DEFAULT_SECTION {
			sections = append(sections[:i], sections[i+1:]...)
		}
	}
	return sections, nil
}

func SetAwsCredentialsFromSection(opts *options.Options, section *ini.Section) error {
	var err error
	opts.AwsSecret.AccessKeyId, err = getValueByKey(section, "aws_access_key_id")
	if err != nil {
		return err
	}
	opts.AwsSecret.SecretAccessKey, err = getValueByKey(section, "aws_secret_access_key")
	if err != nil {
		return err
	}
	return nil
}

func getValueByKey(section *ini.Section, key string) (string, error) {
	if accessKeyId := section.Key(key); accessKeyId == nil || accessKeyId.Value() == "" {
		return "", errors.Errorf("No value for key '%s' in section [%s] of credentials file", key, section.Name())
	} else {
		return accessKeyId.Value(), nil
	}
}
