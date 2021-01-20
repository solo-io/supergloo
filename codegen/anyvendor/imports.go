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

	envoyAccessLog(anyVendorImports)

	return anyVendorImports
}

// import Envoy's accesslog.proto file and all required transitive dependencies
func envoyAccessLog(anyVendorImports *skv2_anyvendor.Imports) {
	anyVendorImports.External["github.com/envoyproxy/data-plane-api"] = []string{
		"envoy/data/accesslog/v3/accesslog.proto",
		"envoy/config/route/v3/route_components.proto",
		"envoy/config/core/v3/extension.proto",
		"envoy/config/core/v3/config_source.proto",
		"envoy/config/core/v3/grpc_service.proto",
		"envoy/config/core/v3/proxy_protocol.proto",
		"envoy/config/core/v3/base.proto",
		"envoy/config/core/v3/address.proto",
		"envoy/config/core/v3/socket_option.proto",
		"envoy/config/core/v3/backoff.proto",
		"envoy/config/core/v3/http_uri.proto",
		"envoy/type/v3/percent.proto",
		"envoy/type/v3/semantic_version.proto",
		"envoy/annotations/deprecation.proto",
		"envoy/type/matcher/v3/regex.proto",
		"envoy/type/matcher/v3/string.proto",
		"envoy/type/matcher/v3/metadata.proto",
		"envoy/type/matcher/v3/value.proto",
		"envoy/type/matcher/v3/number.proto",
		"envoy/type/metadata/v3/metadata.proto",
		"envoy/type/tracing/v3/custom_tag.proto",
		"envoy/type/v3/range.proto",
	}

	anyVendorImports.External["github.com/cncf/udpa"] = []string{
		"udpa/annotations/sensitive.proto",
		"udpa/annotations/versioning.proto",
		"udpa/annotations/migrate.proto",
		"udpa/annotations/status.proto",
		"xds/core/v3/authority.proto",
	}

	anyVendorImports.External["github.com/envoyproxy/protoc-gen-validate"] = []string{
		"validate/validate.proto",
	}
}
