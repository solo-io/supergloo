module github.com/solo-io/mesh-projects

go 1.13

require (
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/go-test/deep v1.0.3 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.0
	github.com/golang/protobuf v1.3.2
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/wire v0.4.0
	github.com/gophercloud/gophercloud v0.6.0 // indirect
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/hashicorp/consul v1.6.2
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/hcl v1.0.0
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mattn/go-shellwords v1.0.7
	github.com/olekukonko/tablewriter v0.0.0-20170122224234-a0225b3f23b5
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pseudomuto/protoc-gen-doc v1.3.0
	github.com/pseudomuto/protokit v0.2.0
	github.com/rotisserie/eris v0.2.0
	github.com/sergi/go-diff v1.1.0
	github.com/servicemeshinterface/smi-sdk-go v0.3.0
	github.com/solo-io/autopilot v0.1.1-0.20200214195428-8aeb08a8feca
	github.com/solo-io/go-utils v0.15.1
	github.com/solo-io/protoc-gen-ext v0.0.7
	github.com/solo-io/reporting-client v0.1.3
	github.com/solo-io/solo-kit v0.13.3
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.13.0
	golang.org/x/sys v0.0.0-20200120151820-655fe14d7479 // indirect
	google.golang.org/genproto v0.0.0-20200117163144-32f20d992d24 // indirect
	helm.sh/helm/v3 v3.0.0
	istio.io/api v0.0.0-20200218013118-7fd43ea7fc2b
	istio.io/client-go v0.0.0-20200218195608-60c238c92aa0
	istio.io/istio v0.0.0-20200215010343-d9274c558175
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.17.4
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	github.com/solo-io/autopilot => github.com/solo-io/autopilot v0.1.1-0.20200214195428-8aeb08a8feca

	k8s.io/api => k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.2
	k8s.io/apiserver => k8s.io/apiserver v0.17.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.2
	k8s.io/client-go => k8s.io/client-go v0.17.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.2
	k8s.io/code-generator => k8s.io/code-generator v0.17.2
	k8s.io/component-base => k8s.io/component-base v0.17.2
	k8s.io/cri-api => k8s.io/cri-api v0.17.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.2
	k8s.io/heapster => k8s.io/heapster v1.17.1
	k8s.io/klog => github.com/stefanprodan/klog v0.0.0-20190418165334-9cbb78b20423
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.2
	k8s.io/kubectl => k8s.io/kubectl v0.17.2
	k8s.io/kubelet => k8s.io/kubelet v0.17.2
	k8s.io/kubernetes => k8s.io/kubernetes v0.17.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.2
	k8s.io/metrics => k8s.io/metrics v0.17.2
	k8s.io/node-api => k8s.io/node-api v0.17.2
	k8s.io/repo-infra => k8s.io/repo-infra v0.17.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.2
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.2
	k8s.io/sample-controller => k8s.io/sample-controller v0.17.2
)
