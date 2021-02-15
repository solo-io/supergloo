module github.com/solo-io/gloo-mesh

go 1.14

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884
	k8s.io/api => k8s.io/api v0.19.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.7
	k8s.io/client-go => k8s.io/client-go v0.19.7
	k8s.io/kubectl => k8s.io/kubectl v0.19.7

)

require (
	cloud.google.com/go/logging v1.0.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9 // indirect
	fortio.org/fortio v1.4.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/alicebob/gopher-json v0.0.0-20180125190556-5a6b3ba71ee6 // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible // indirect
	github.com/aws/aws-app-mesh-controller-for-k8s v1.1.1
	github.com/aws/aws-sdk-go v1.36.19
	github.com/cactus/go-statsd-client v3.1.1+incompatible // indirect
	github.com/circonus-labs/circonusllhist v0.1.4 // indirect
	github.com/cncf/udpa/go v0.0.0-20201120205902-5459f2c99403
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/dchest/siphash v1.1.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1 // indirect
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210115003313-31f9241a16e6
	github.com/fluent/fluent-logger-golang v1.3.0 // indirect
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-logr/zapr v0.3.0 // indirect
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.6 // indirect
	github.com/go-redis/redis v6.10.2+incompatible // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f // indirect
	github.com/gomodule/redigo v1.8.0 // indirect
	github.com/google/cel-go v0.4.1 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20171214222146-0e7658f8ee99 // indirect
	github.com/hashicorp/consul v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/howeyc/fsnotify v0.9.0 // indirect
	github.com/iancoleman/strcase v0.1.3
	github.com/kr/pretty v0.2.1 // indirect
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/mholt/archiver v3.1.1+incompatible // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.2
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.4
	github.com/open-policy-agent/opa v0.8.2 // indirect
	github.com/openservicemesh/osm v0.3.0
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180306154005-525d0eb5f91d // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/prometheus/prom2json v1.2.2 // indirect
	github.com/pseudomuto/protoc-gen-doc v1.3.2
	github.com/pseudomuto/protokit v0.2.0
	github.com/rotisserie/eris v0.4.0
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/sirupsen/logrus v1.7.0
	github.com/solo-io/anyvendor v0.0.3
	github.com/solo-io/external-apis v0.1.4
	github.com/solo-io/go-list-licenses v0.1.0
	github.com/solo-io/go-utils v0.20.4
	github.com/solo-io/k8s-utils v0.0.3
	github.com/solo-io/protoc-gen-ext v0.0.15
	github.com/solo-io/skv2 v0.17.4
	github.com/solo-io/solo-kit v0.16.0
	github.com/spf13/afero v1.5.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/tinylib/msgp v1.0.2 // indirect
	github.com/tv42/httpunix v0.0.0-20191220191345-2ba4b9c3382c // indirect
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	github.com/yashtewari/glob-intersection v0.0.0-20180206001645-7af743e8ec84 // indirect
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb // indirect
	go.opencensus.io v0.22.5 // indirect
	go.uber.org/atomic v1.7.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0
	gopkg.in/d4l3k/messagediff.v1 v1.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.4.2
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	istio.io/api v0.0.0-20210121191246-a7d07ed40d49
	istio.io/client-go v0.0.0-20200908160912-f99162621a1a
	istio.io/istio v0.0.0-20210126155301-c18b82ae7269
	k8s.io/api v0.20.1
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	k8s.io/kubernetes v1.13.0
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime v0.7.0
)
