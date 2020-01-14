package translator

import (
	"github.com/rotisserie/eris"
)

var (
	FailedToGenerateRenderValues = func(err error, namespace, name string) error {
		return eris.Errorf("Failed to generate render arguments from app state %v.%v", namespace, name)
	}

	FlavorNotFoundError = func(name, version string) error {
		return eris.Errorf("Could not find flavor with name %v for version %v", name, version)
	}

	UnsupportedParamType = func(paramValueType interface{}) error {
		return eris.Errorf("The operator does not currently support params of type %T", paramValueType)
	}
)
