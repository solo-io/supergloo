module github.com/solo-io/gloo-mesh

go 1.14

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// https://github.com/ory/dockertest/issues/208#issuecomment-686820414
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6

	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8

	k8s.io/client-go => k8s.io/client-go v0.18.8
)

require (
	cloud.google.com/go v0.66.0 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/aws/aws-app-mesh-controller-for-k8s v1.1.1
	github.com/aws/aws-sdk-go v1.33.11
	github.com/cncf/udpa v0.0.2-0.20201211205326-cc1b757b3edd
	github.com/cncf/udpa/go v0.0.0-20201120205902-5459f2c99403
	github.com/docker/distribution v2.7.1+incompatible
	github.com/envoyproxy/data-plane-api v0.0.0-20210110221212-8c6115ddd1b8
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210110222040-c850101e02b6
	github.com/envoyproxy/protoc-gen-validate v0.4.1 // indirect
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/go-logr/zapr v0.3.0
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.6 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/google/go-github v17.0.0+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.2
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/openservicemesh/osm v0.3.0
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/pseudomuto/protoc-gen-doc v1.3.2
	github.com/pseudomuto/protokit v0.2.0
	github.com/rotisserie/eris v0.4.0
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/sirupsen/logrus v1.6.0
	github.com/solo-io/anyvendor v0.0.1
	github.com/solo-io/external-apis v0.1.1
	github.com/solo-io/go-utils v0.20.0
	github.com/solo-io/k8s-utils v0.0.3
	github.com/solo-io/protoc-gen-ext v0.0.13
	github.com/solo-io/skv2 v0.15.4-0.20210111180817-82fcdc310655
	github.com/solo-io/solo-kit v0.16.0
	github.com/spf13/afero v1.3.4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/atomic v1.6.0
	go.uber.org/zap v1.15.0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/tools v0.0.0-20200928201943-a0ef9b62deab // indirect
	google.golang.org/grpc v1.31.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.2.4
	istio.io/api v0.0.0-20210107192247-a205c627e4b9
	istio.io/client-go v0.0.0-20200812230733-f5504d568313
	istio.io/gogo-genproto v0.0.0-20210107204948-697d6f912366 // indirect
	istio.io/istio v0.0.0-20200821180042-b0e61d10cbae
	k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/klog/v2 v2.2.0 // indirect
	k8s.io/kubernetes v1.13.0
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime v0.6.2
)
