package linkerd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/linkerd/linkerd2/pkg/k8s"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/errors"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
)

const chartPath_stable221 = "https://storage.googleapis.com/supergloo-charts/linkerd-stable-2.2.1.tgz"

func (o *installOpts) chartURI() (string, error) {
	switch o.installVersion {
	case Version_stable221:
		return chartPath_stable221, nil
	}
	return "", errors.Errorf("version %v is not a supported linkerd version. supported: %v", o.installVersion, supportedVersions)
}

func (o *installOpts) values() (string, error) {
	opts := newCliInstallOptions()
	opts.proxyAutoInject = o.enableAutoInject
	if o.enableMtls {
		opts.tls = optionalTLS
	}
	valuesCfg, err := validateAndBuildConfig(opts, o.installNamespace)
	if err != nil {
		return "", err
	}
	rawYaml, err := yaml.Marshal(valuesCfg)
	if err != nil {
		return "", err
	}
	return string(rawYaml), nil
}

// the code below mostly based on https://github.com/linkerd/linkerd2/blob/release/stable-2.2/cli/cmd/install.go#L24
// with some modifications
// defaults are fine for most of this,
// we can open this up to more configuration in the future
// currently only autoinjection and mtls support are
// exposed
type helmValues struct {
	Namespace                        string
	ControllerImage                  string
	WebImage                         string
	PrometheusImage                  string
	PrometheusVolumeName             string
	GrafanaImage                     string
	GrafanaVolumeName                string
	ControllerReplicas               uint
	ImagePullPolicy                  string
	UUID                             string
	CliVersion                       string
	ControllerLogLevel               string
	ControllerComponentLabel         string
	CreatedByAnnotation              string
	ProxyAPIPort                     uint
	EnableTLS                        bool
	TLSTrustAnchorVolumeName         string
	TLSSecretsVolumeName             string
	TLSTrustAnchorConfigMapName      string
	ProxyContainerName               string
	TLSTrustAnchorFileName           string
	TLSCertFileName                  string
	TLSPrivateKeyFileName            string
	TLSTrustAnchorVolumeSpecFileName string
	TLSIdentityVolumeSpecFileName    string
	InboundPort                      uint
	OutboundPort                     uint
	IgnoreInboundPorts               string
	IgnoreOutboundPorts              string
	InboundAcceptKeepaliveMs         uint
	OutboundConnectKeepaliveMs       uint
	ProxyAutoInjectEnabled           bool
	ProxyInjectAnnotation            string
	ProxyInjectDisabled              string
	ProxyLogLevel                    string
	ProxyUID                         int64
	ProxyMetricsPort                 uint
	ProxyControlPort                 uint
	ProxySpecFileName                string
	ProxyInitSpecFileName            string
	ProxyInitImage                   string
	ProxyImage                       string
	ProxyResourceRequestCPU          string
	ProxyResourceRequestMemory       string
	SingleNamespace                  bool
	EnableHA                         bool
	ControllerUID                    int64
	ProfileSuffixes                  string
	EnableH2Upgrade                  bool
	NoInitContainer                  bool
}

const (
	defaultControllerReplicas   = 1
	defaultHAControllerReplicas = 3
)

func newCliInstallOptions() *cliInstallOptions {
	return &cliInstallOptions{
		controllerReplicas: defaultControllerReplicas,
		controllerLogLevel: "info",
		proxyAutoInject:    false,
		singleNamespace:    false,
		highAvailability:   false,
		controllerUID:      2103,
		disableH2Upgrade:   false,
		proxyConfigOptions: newProxyConfigOptions(),
	}
}
func validateAndBuildConfig(options *cliInstallOptions, installNamespace string) (*helmValues, error) {
	if err := options.validate(); err != nil {
		return nil, err
	}

	ignoreInboundPorts := []string{
		fmt.Sprintf("%d", options.proxyControlPort),
		fmt.Sprintf("%d", options.proxyMetricsPort),
	}
	for _, p := range options.ignoreInboundPorts {
		ignoreInboundPorts = append(ignoreInboundPorts, fmt.Sprintf("%d", p))
	}
	ignoreOutboundPorts := []string{}
	for _, p := range options.ignoreOutboundPorts {
		ignoreOutboundPorts = append(ignoreOutboundPorts, fmt.Sprintf("%d", p))
	}

	if options.highAvailability && options.controllerReplicas == defaultControllerReplicas {
		options.controllerReplicas = defaultHAControllerReplicas
	}

	if options.highAvailability && options.proxyCPURequest == "" {
		options.proxyCPURequest = "10m"
	}

	if options.highAvailability && options.proxyMemoryRequest == "" {
		options.proxyMemoryRequest = "20Mi"
	}

	profileSuffixes := "."
	if options.proxyConfigOptions.disableExternalProfiles {
		profileSuffixes = "svc.cluster.local."
	}

	return &helmValues{
		Namespace:                        installNamespace,
		ControllerImage:                  fmt.Sprintf("%s/controller:%s", options.dockerRegistry, options.linkerdVersion),
		WebImage:                         fmt.Sprintf("%s/web:%s", options.dockerRegistry, options.linkerdVersion),
		PrometheusImage:                  "prom/prometheus:v2.4.0",
		PrometheusVolumeName:             "data",
		GrafanaImage:                     fmt.Sprintf("%s/grafana:%s", options.dockerRegistry, options.linkerdVersion),
		GrafanaVolumeName:                "data",
		ControllerReplicas:               options.controllerReplicas,
		ImagePullPolicy:                  options.imagePullPolicy,
		UUID:                             uuid.NewV4().String(),
		CliVersion:                       k8s.CreatedByAnnotationValue(),
		ControllerLogLevel:               options.controllerLogLevel,
		ControllerComponentLabel:         k8s.ControllerComponentLabel,
		ControllerUID:                    options.controllerUID,
		CreatedByAnnotation:              k8s.CreatedByAnnotation,
		ProxyAPIPort:                     options.proxyAPIPort,
		EnableTLS:                        options.enableTLS(),
		TLSTrustAnchorVolumeName:         k8s.TLSTrustAnchorVolumeName,
		TLSSecretsVolumeName:             k8s.TLSSecretsVolumeName,
		TLSTrustAnchorConfigMapName:      k8s.TLSTrustAnchorConfigMapName,
		ProxyContainerName:               k8s.ProxyContainerName,
		TLSTrustAnchorFileName:           k8s.TLSTrustAnchorFileName,
		TLSCertFileName:                  k8s.TLSCertFileName,
		TLSPrivateKeyFileName:            k8s.TLSPrivateKeyFileName,
		TLSTrustAnchorVolumeSpecFileName: k8s.TLSTrustAnchorVolumeSpecFileName,
		TLSIdentityVolumeSpecFileName:    k8s.TLSIdentityVolumeSpecFileName,
		InboundPort:                      options.inboundPort,
		OutboundPort:                     options.outboundPort,
		IgnoreInboundPorts:               strings.Join(ignoreInboundPorts, ","),
		IgnoreOutboundPorts:              strings.Join(ignoreOutboundPorts, ","),
		InboundAcceptKeepaliveMs:         defaultKeepaliveMs,
		OutboundConnectKeepaliveMs:       defaultKeepaliveMs,
		ProxyAutoInjectEnabled:           options.proxyAutoInject,
		ProxyInjectAnnotation:            k8s.ProxyInjectAnnotation,
		ProxyInjectDisabled:              k8s.ProxyInjectDisabled,
		ProxyLogLevel:                    options.proxyLogLevel,
		ProxyUID:                         options.proxyUID,
		ProxyMetricsPort:                 options.proxyMetricsPort,
		ProxyControlPort:                 options.proxyControlPort,
		ProxySpecFileName:                k8s.ProxySpecFileName,
		ProxyInitSpecFileName:            k8s.ProxyInitSpecFileName,
		ProxyInitImage:                   options.taggedProxyInitImage(),
		ProxyImage:                       options.taggedProxyImage(),
		ProxyResourceRequestCPU:          options.proxyCPURequest,
		ProxyResourceRequestMemory:       options.proxyMemoryRequest,
		SingleNamespace:                  options.singleNamespace,
		EnableHA:                         options.highAvailability,
		ProfileSuffixes:                  profileSuffixes,
		EnableH2Upgrade:                  !options.disableH2Upgrade,
		NoInitContainer:                  options.noInitContainer,
	}, nil
}

// cliInstallOptions holds values for command line flags that apply to the install
// command. All fields in this struct should have corresponding flags added in
// the newCmdInstall func later in this file. It also embeds proxyConfigOptions
// in order to hold values for command line flags that apply to both inject and
// install.
type cliInstallOptions struct {
	controllerReplicas uint
	controllerLogLevel string
	proxyAutoInject    bool
	singleNamespace    bool
	highAvailability   bool
	controllerUID      int64
	disableH2Upgrade   bool
	*proxyConfigOptions
}

func (options *cliInstallOptions) validate() error {
	if _, err := log.ParseLevel(options.controllerLogLevel); err != nil {
		return fmt.Errorf("--controller-log-level must be one of: panic, fatal, error, warn, info, debug")
	}

	if options.proxyAutoInject && options.singleNamespace {
		return fmt.Errorf("The --proxy-auto-inject and --single-namespace flags cannot both be specified together")
	}

	return options.proxyConfigOptions.validate()
}

// proxyConfigOptions holds values for command line flags that apply to both the
// install and inject commands. All fields in this struct should have
// corresponding flags added in the addProxyConfigFlags func later in this file.
type proxyConfigOptions struct {
	linkerdVersion          string
	proxyImage              string
	initImage               string
	dockerRegistry          string
	imagePullPolicy         string
	inboundPort             uint
	outboundPort            uint
	ignoreInboundPorts      []uint
	ignoreOutboundPorts     []uint
	proxyUID                int64
	proxyLogLevel           string
	proxyAPIPort            uint
	proxyControlPort        uint
	proxyMetricsPort        uint
	proxyCPURequest         string
	proxyMemoryRequest      string
	tls                     string
	disableExternalProfiles bool
	noInitContainer         bool

	// proxyOutboundCapacity is a special case that's only used for injecting the
	// proxy into the control plane install, and as such it does not have a
	// corresponding command line flag.
	proxyOutboundCapacity map[string]uint
}

func (options *proxyConfigOptions) enableTLS() bool {
	return options.tls == optionalTLS
}

func (options *proxyConfigOptions) taggedProxyImage() string {
	image := strings.Replace(options.proxyImage, defaultDockerRegistry, options.dockerRegistry, 1)
	return fmt.Sprintf("%s:%s", image, options.linkerdVersion)
}

func (options *proxyConfigOptions) taggedProxyInitImage() string {
	image := strings.Replace(options.initImage, defaultDockerRegistry, options.dockerRegistry, 1)
	return fmt.Sprintf("%s:%s", image, options.linkerdVersion)
}

var (
	// These regexs are not as strict as they could be, but are a quick and dirty
	// sanity check against illegal characters.
	alphaNumDashDot           = regexp.MustCompile("^[\\.a-zA-Z0-9-]+$")
	alphaNumDashDotSlashColon = regexp.MustCompile("^[\\./a-zA-Z0-9-:]+$")
)

func (options *proxyConfigOptions) validate() error {
	if !alphaNumDashDot.MatchString(options.linkerdVersion) {
		return fmt.Errorf("%s is not a valid version", options.linkerdVersion)
	}

	if !alphaNumDashDotSlashColon.MatchString(options.dockerRegistry) {
		return fmt.Errorf("%s is not a valid Docker registry. The url can contain only letters, numbers, dash, dot, slash and colon", options.dockerRegistry)
	}

	if options.imagePullPolicy != "Always" && options.imagePullPolicy != "IfNotPresent" && options.imagePullPolicy != "Never" {
		return fmt.Errorf("--image-pull-policy must be one of: Always, IfNotPresent, Never")
	}

	if options.proxyCPURequest != "" {
		if _, err := k8sResource.ParseQuantity(options.proxyCPURequest); err != nil {
			return fmt.Errorf("Invalid cpu request '%s' for --proxy-cpu flag", options.proxyCPURequest)
		}
	}

	if options.proxyMemoryRequest != "" {
		if _, err := k8sResource.ParseQuantity(options.proxyMemoryRequest); err != nil {
			return fmt.Errorf("Invalid memory request '%s' for --proxy-memory flag", options.proxyMemoryRequest)
		}
	}

	if options.tls != "" && options.tls != optionalTLS {
		return fmt.Errorf("--tls must be blank or set to \"%s\"", optionalTLS)
	}

	return nil
}

const (
	optionalTLS           = "optional"
	defaultDockerRegistry = "gcr.io/linkerd-io"
	defaultKeepaliveMs    = 10000
)

func newProxyConfigOptions() *proxyConfigOptions {
	return &proxyConfigOptions{
		linkerdVersion:          "stable-2.2.1",
		proxyImage:              defaultDockerRegistry + "/proxy",
		initImage:               defaultDockerRegistry + "/proxy-init",
		dockerRegistry:          defaultDockerRegistry,
		imagePullPolicy:         "IfNotPresent",
		inboundPort:             4143,
		outboundPort:            4140,
		ignoreInboundPorts:      nil,
		ignoreOutboundPorts:     nil,
		proxyUID:                2102,
		proxyLogLevel:           "warn,linkerd2_proxy=info",
		proxyAPIPort:            8086,
		proxyControlPort:        4190,
		proxyMetricsPort:        4191,
		proxyCPURequest:         "",
		proxyMemoryRequest:      "",
		tls:                     "",
		disableExternalProfiles: false,
		noInitContainer:         false,
		proxyOutboundCapacity:   map[string]uint{},
	}
}
