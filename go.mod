module github.com/solo-io/service-mesh-hub

go 1.14

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0

	k8s.io/client-go => k8s.io/client-go v0.18.6

)

require (
	cloud.google.com/go/logging v1.0.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9 // indirect
	contrib.go.opencensus.io/exporter/zipkin v0.1.1 // indirect
	github.com/Azure/go-autorest/autorest v0.9.4 // indirect
	github.com/Jeffail/gabs v1.1.0 // indirect
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/SAP/go-hdb v0.12.0 // indirect
	github.com/SermoDigital/jose v0.0.0-20180104203859-803625baeddc // indirect
	github.com/StackExchange/wmi v0.0.0-20180116203802-5d049714c4a6 // indirect
	github.com/abdullin/seq v0.0.0-20160510034733-d5467c17e7af // indirect
	github.com/alicebob/gopher-json v0.0.0-20180125190556-5a6b3ba71ee6 // indirect
	github.com/alicebob/miniredis v0.0.0-20180201100744-9d52b1fc8da9 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
	github.com/cactus/go-statsd-client v3.1.1+incompatible // indirect
	github.com/circonus-labs/circonusllhist v0.1.4 // indirect
	github.com/cncf/udpa/go v0.0.0-20200629203442-efcf912fb354
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/coredns/coredns v1.1.2 // indirect
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/dchest/siphash v1.1.0 // indirect
	github.com/digitalocean/godo v1.10.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/duosecurity/duo_api_golang v0.0.0-20190308151101-6c680f768e74 // indirect
	github.com/elazarl/go-bindata-assetfs v0.0.0-20160803192304-e1a2a7ec64b0 // indirect
	github.com/envoyproxy/go-control-plane v0.9.7-0.20200730005029-803dd64f0468
	github.com/fluent/fluent-logger-golang v1.3.0 // indirect
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-ldap/ldap v3.0.2+incompatible // indirect
	github.com/go-logr/zapr v0.1.1
	github.com/go-ole/go-ole v1.2.1 // indirect
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.6 // indirect
	github.com/go-redis/redis v6.10.2+incompatible // indirect
	github.com/gocql/gocql v0.0.0-20180617115710-e06f8c1bcd78 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f // indirect
	github.com/google/cel-go v0.2.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20171214222146-0e7658f8ee99 // indirect
	github.com/hashicorp/go-bexpr v0.1.2 // indirect
	github.com/hashicorp/go-checkpoint v0.0.0-20171009173528-1545e56e46de // indirect
	github.com/hashicorp/go-memdb v0.0.0-20180223233045-1289e7fffe71 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-raftchunking v0.6.1 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hil v0.0.0-20160711231837-1e86c6b523c5 // indirect
	github.com/hashicorp/mdns v1.0.1 // indirect
	github.com/hashicorp/net-rpc-msgpackrpc v0.0.0-20151116020338-a14192a58a69 // indirect
	github.com/hashicorp/vic v1.5.1-0.20190403131502-bbfe86ec9443 // indirect
	github.com/jarcoal/httpmock v0.0.0-20180424175123-9c70cfe4a1da // indirect
	github.com/jefferai/jsonx v0.0.0-20160721235117-9cc31c3135ee // indirect
	github.com/joyent/triton-go v0.0.0-20180628001255-830d2b111e62 // indirect
	github.com/keybase/go-crypto v0.0.0-20180614160407-5114a9a81e1b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/mholt/archiver v3.1.1+incompatible // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/nicolai86/scaleway-sdk v1.10.2-0.20180628010248-798f60e20bb2 // indirect
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.2
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/open-policy-agent/opa v0.8.2 // indirect
	github.com/openshift/api v3.9.1-0.20191008181517-e4fd21196097+incompatible // indirect
	github.com/openzipkin/zipkin-go v0.1.7 // indirect
	github.com/ory/dockertest v3.3.4+incompatible // indirect
	github.com/packethost/packngo v0.1.1-0.20180711074735-b9cb5096f54c // indirect
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180306154005-525d0eb5f91d // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/prom2json v1.2.2 // indirect
	github.com/renier/xmlrpc v0.0.0-20170708154548-ce4a1a486c03 // indirect
	github.com/rotisserie/eris v0.4.0
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/shirou/gopsutil v0.0.0-20181107111621-48177ef5f880 // indirect
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/softlayer/softlayer-go v0.0.0-20180806151055-260589d94c7d // indirect
	github.com/solo-io/external-apis v0.0.6
	github.com/solo-io/go-utils v0.17.0
	github.com/solo-io/skv2 v0.8.1
	github.com/solo-io/solo-kit v0.14.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/tent/http-link-go v0.0.0-20130702225549-ac974c61c2f9 // indirect
	github.com/tinylib/msgp v1.0.2 // indirect
	github.com/tv42/httpunix v0.0.0-20191220191345-2ba4b9c3382c // indirect
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	github.com/vmware/govmomi v0.18.0 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/yashtewari/glob-intersection v0.0.0-20180206001645-7af743e8ec84 // indirect
	github.com/yuin/gopher-lua v0.0.0-20180316054350-84ea3a3c79b3 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.15.0
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20160818020120-3f83fa500528 // indirect
	gopkg.in/ory-am/dockertest.v3 v3.3.4 // indirect
	helm.sh/helm/v3 v3.2.4
	istio.io/api v0.0.0-20200808213952-0bb7e74acfe8
	istio.io/client-go v0.0.0-20200807223845-61c70ad04ec9
	istio.io/istio v0.0.0-20200812162523-97f8edc96a95
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/helm v2.14.3+incompatible // indirect
	k8s.io/kube-aggregator v0.18.6 // indirect
	rsc.io/quote/v3 v3.1.0 // indirect
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
)

replace github.com/solo-io/skv2 => ../skv2

replace github.com/solo-io/external-apis => ../external-apis
