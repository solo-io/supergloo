package anyvendor

import "github.com/solo-io/skv2/codegen/skv2_anyvendor"

func AnyVendorImports() *skv2_anyvendor.Imports {
	anyVendorImports := skv2_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})

	anyVendorImports.External["github.com/solo-io/skv2"] = []string{
		"api/core/v1/*.proto",
		"crds/multicluster.solo.io_v1alpha1_crds.yaml",
	}

	anyVendorImports.External["istio.io/api"] = []string{
		"networking/v1alpha3/*.proto",
		"common-protos/google/api/field_behavior.proto",
	}

	anyVendorImports.External["k8s.io/apimachinery"] = []string{
		"pkg/apis/meta/v1/generated.proto",
		"pkg/runtime/generated.proto",
		"pkg/runtime/schema/generated.proto",
	}
	return anyVendorImports
}
