module github.com/solo-io/service-mesh-hub

go 1.14

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.9.5

	// Required for proper serialization of CRDs
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.5

	// https://github.com/ory/dockertest/issues/208#issuecomment-686820414
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6

	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884

	k8s.io/client-go => k8s.io/client-go v0.18.6
)

require (
	cloud.google.com/go v0.66.0 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/aws/aws-app-mesh-controller-for-k8s v1.1.1
	github.com/aws/aws-sdk-go v1.33.11
	github.com/cncf/udpa/go v0.0.0-20200629203442-efcf912fb354
	github.com/docker/distribution v2.7.1+incompatible
	github.com/envoyproxy/go-control-plane v0.9.7-0.20200811182123-112a4904c4b0
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-logr/zapr v0.1.1
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.6 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.2
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/openservicemesh/osm v0.3.0
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/pseudomuto/protoc-gen-doc v1.3.0
	github.com/pseudomuto/protokit v0.2.0
	github.com/rotisserie/eris v0.4.0
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/sirupsen/logrus v1.6.0
	github.com/solo-io/anyvendor v0.0.1
	github.com/solo-io/external-apis v0.1.1
	github.com/solo-io/go-utils v0.18.1
	github.com/solo-io/skv2 v0.13.0
	github.com/solo-io/solo-kit v0.14.0
	github.com/spf13/afero v1.3.4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.15.0
	golang.org/x/tools v0.0.0-20200928201943-a0ef9b62deab // indirect
	helm.sh/helm/v3 v3.2.4
	istio.io/api v0.0.0-20200819225923-c78f387f78a2
	istio.io/client-go v0.0.0-20200812230733-f5504d568313
	istio.io/istio v0.0.0-20200821180042-b0e61d10cbae
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v8.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.2
)
