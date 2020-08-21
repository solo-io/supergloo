module github.com/solo-io/service-mesh-hub

go 1.14

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	github.com/golang/protobuf => github.com/golang/protobuf v1.3.5
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884

	github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.9.5

	k8s.io/client-go => k8s.io/client-go v0.18.6
)

require (
	cloud.google.com/go v0.61.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.2.0 // indirect
	fortio.org/fortio v1.6.3 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.33.11 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cheggaaa/pb/v3 v3.0.4 // indirect
	github.com/cncf/udpa/go v0.0.0-20200629203442-efcf912fb354
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/containernetworking/cni v0.7.0-alpha1 // indirect
	github.com/containernetworking/plugins v0.7.3 // indirect
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.9.7-0.20200811182123-112a4904c4b0
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-logr/zapr v0.1.1
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.6 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f // indirect
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lestrrat-go/jwx v1.0.3 // indirect
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/mholt/archiver v3.1.1+incompatible // indirect
	github.com/miekg/dns v1.1.30 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.2
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/openservicemesh/osm v0.3.0
	github.com/openshift/api v0.0.0-20200713203337-b2494ecb17dd // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/pseudomuto/protoc-gen-doc v1.3.0
	github.com/pseudomuto/protokit v0.2.0
	github.com/rotisserie/eris v0.4.0
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/sirupsen/logrus v1.6.0
	github.com/solo-io/anyvendor v0.0.1
	github.com/solo-io/external-apis v0.1.0
	github.com/solo-io/go-utils v0.17.0
	github.com/solo-io/skv2 v0.10.0
	github.com/solo-io/solo-kit v0.14.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/yl2chen/cidranger v1.0.0 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/genproto v0.0.0-20200722002428-88e341933a54 // indirect
	google.golang.org/grpc/examples v0.0.0-20200818224027-0f73133e3aa3 // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	helm.sh/helm/v3 v3.2.4
	istio.io/api v0.0.0-20200819225923-c78f387f78a2
	istio.io/client-go v0.0.0-20200812230733-f5504d568313
	istio.io/gogo-genproto v0.0.0-20200720193312-b523a30fe746 // indirect
	istio.io/istio v0.0.0-20200821180042-b0e61d10cbae
	istio.io/pkg v0.0.0-20200721143030-6b837ddaf2ab // indirect
	k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver v0.18.8 // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/kubectl v0.18.8 // indirect
	k8s.io/utils v0.0.0-20200720150651-0bdb4ca86cbc // indirect
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/service-apis v0.0.0-20200731055707-56154e7bfde5 // indirect
)
