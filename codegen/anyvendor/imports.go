package anyvendor

import (
	"github.com/solo-io/skv2/codegen/skv2_anyvendor"
)

func AnyVendorImports() *skv2_anyvendor.Imports {
	anyVendorImports := skv2_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})

	anyVendorImports.External["github.com/solo-io/skv2"] = []string{
		"api/core/v1/*.proto",
		"crds/multicluster.solo.io_v1alpha1_crds.yaml",
	}

	anyVendorImports.External["github.com/gogo/protobuf"] = []string{
		"gogoproto/*.proto",
	}

	anyVendorImports.External["istio.io/api"] = []string{
		"networking/v1alpha3/*.proto",
		"common-protos/google/api/field_behavior.proto",
	}

	// used for rate limit server config
	anyVendorImports.External["github.com/solo-io/solo-apis"] = []string{
		"api/rate-limiter/v1alpha1/ratelimit.proto",
		"crds/ratelimit.solo.io_v1alpha1_crds.yaml",
	}

	// used for a proto option which disables openapi validation on fields
	anyVendorImports.External["cuelang.org/go"] = []string{
		"encoding/protobuf/cue/cue.proto",
	}

	return anyVendorImports
}
