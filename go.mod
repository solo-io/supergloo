module github.com/solo-io/service-mesh-hub

go 1.14

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	k8s.io/client-go => k8s.io/client-go v0.18.6

)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/cncf/udpa/go v0.0.0-20200629203442-efcf912fb354
	github.com/docker/distribution v2.7.1+incompatible
	github.com/envoyproxy/go-control-plane v0.9.7-0.20200730005029-803dd64f0468
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-logr/zapr v0.1.1
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/olekukonko/tablewriter v0.0.2
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/openservicemesh/osm v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/rotisserie/eris v0.4.0
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/sirupsen/logrus v1.6.0
	github.com/solo-io/external-apis v0.0.6
	github.com/solo-io/go-utils v0.17.0
	github.com/solo-io/skv2 v0.10.0
	github.com/solo-io/solo-kit v0.14.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.15.0
	helm.sh/helm/v3 v3.2.4
	istio.io/api v0.0.0-20200808213952-0bb7e74acfe8
	istio.io/client-go v0.0.0-20200807223845-61c70ad04ec9
	istio.io/istio v0.0.0-20200812162523-97f8edc96a95
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v8.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.2
)

replace github.com/solo-io/external-apis => ../external-apis
