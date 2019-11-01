package translator

import (
	"github.com/solo-io/go-utils/errors"
)

var (
	FailedToGenerateRenderValues = func(err error, namespace, name string) error {
		return errors.Errorf("Failed to generate render arguments from app state %v.%v", namespace, name)
	}

	FlavorNotFoundError = func(name, version string) error {
		return errors.Errorf("Could not find flavor with name %v for version %v")
	}

	UnsupportedParamType = func(paramValueType interface{}) error {
		return errors.Errorf("The operator does not currently support params of type %T", paramValueType)
	}
)
