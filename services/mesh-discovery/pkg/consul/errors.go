package consul

import "github.com/rotisserie/eris"

var (
	InvalidImageFormatError = func(err error, imageName string) error {
		return eris.Wrapf(err, "invalid or unexpected image format for image name: %s", imageName)
	}
	HclParseError = func(err error, invalidHcl string) error {
		return eris.Wrapf(err, "error parsing HCL in consul invocation: %s", invalidHcl)
	}
)
