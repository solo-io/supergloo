package consul

import "github.com/solo-io/go-utils/errors"

var (
	InvalidImageFormatError = func(err error, imageName string) error {
		return errors.Wrapf(err, "invalid or unexpected image format for image name: %s", imageName)
	}
	HclParseError = func(err error, invalidHcl string) error {
		return errors.Wrapf(err, "error parsing HCL in consul invocation: %s", invalidHcl)
	}
)
