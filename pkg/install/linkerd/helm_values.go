package linkerd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	pb "github.com/linkerd/linkerd2/controller/gen/config"
	"github.com/linkerd/linkerd2/pkg/config"
	"github.com/linkerd/linkerd2/pkg/inject"
	"github.com/linkerd/linkerd2/pkg/k8s"
	"github.com/linkerd/linkerd2/pkg/tls"
	"github.com/linkerd/linkerd2/pkg/version"
	log "github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/errors"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	yamlDecoder "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

const chartPath_stable221 = "https://storage.googleapis.com/supergloo-charts/linkerd-stable-2.3.0.tgz"

func (o *installOpts) chartURI() (string, error) {
	switch o.installVersion {
	case Version_stable230:
		return chartPath_stable221, nil
	}
	return "", errors.Errorf("version %v is not a supported linkerd version. supported: %v", o.installVersion, supportedVersions)
}

func (o *installOpts) values() (injector, string, error) {
	opts := newInstallOptionsWithDefaults(o.installNamespace)
	opts.proxyAutoInject = o.enableAutoInject
	if o.enableMtls {
		// cannot currently disable tls in
	}
	values, cfg, err := opts.validateAndBuild()
	if err != nil {
		return injector{}, "", err
	}
	rawYaml, err := yaml.Marshal(values)
	if err != nil {
		return injector{}, "", err
	}
	injector := injector{
		configs: cfg,
		proxyOutboundCapacity: map[string]uint{
			values.PrometheusImage: prometheusProxyOutboundCapacity,
		},
	}
	return injector, string(rawYaml), nil
}

// the code below mostly based on https://github.com/linkerd/linkerd2/blob/release/stable-2.2/cli/cmd/install.go#L24
// with some modifications
// defaults are fine for most of this,
// we can open this up to more configuration in the future
// currently only autoinjection and mtls support are
// exposed
// newInstallOptionsWithDefaults initializes install options with default
// control plane and proxy options.
//
// These options may be overridden on the CLI at install-time and will be
// persisted in Linkerd's control plane configuration to be used at
// injection-time.
type (
	installValues struct {
		Namespace                string
		ControllerImage          string
		WebImage                 string
		PrometheusImage          string
		GrafanaImage             string
		ImagePullPolicy          string
		UUID                     string
		CliVersion               string
		ControllerReplicas       uint
		ControllerLogLevel       string
		PrometheusLogLevel       string
		ControllerComponentLabel string
		CreatedByAnnotation      string
		ProxyContainerName       string
		ProxyAutoInjectEnabled   bool
		ProxyInjectAnnotation    string
		ProxyInjectDisabled      string
		ControllerUID            int64
		EnableH2Upgrade          bool
		NoInitContainer          bool

		Configs configJSONs

		DestinationResources,
		GrafanaResources,
		IdentityResources,
		PrometheusResources,
		ProxyInjectorResources,
		PublicAPIResources,
		SPValidatorResources,
		TapResources,
		WebResources *resources

		Identity *installIdentityValues
	}

	configJSONs struct{ Global, Proxy, Install string }

	resources   struct{ CPU, Memory constraints }
	constraints struct{ Request, Limit string }

	installIdentityValues struct {
		Replicas uint

		TrustDomain     string
		TrustAnchorsPEM string

		Issuer *issuerValues
	}

	issuerValues struct {
		ClockSkewAllowance string
		IssuanceLifetime   string

		KeyPEM, CrtPEM string

		CrtExpiry time.Time

		CrtExpiryAnnotation string
	}

	// installOptions holds values for command line flags that apply to the install
	// command. All fields in this struct should have corresponding flags added in
	// the newCmdInstall func later in this file. It also embeds proxyConfigOptions
	// in order to hold values for command line flags that apply to both inject and
	// install.
	installOptions struct {
		// :custom
		installNamespace   string
		controllerReplicas uint
		controllerLogLevel string
		proxyAutoInject    bool
		highAvailability   bool
		controllerUID      int64
		disableH2Upgrade   bool
		noInitContainer    bool
		identityOptions    *installIdentityOptions
		*proxyConfigOptions

		recordedFlags []*pb.Install_Flag

		// A function pointer that can be overridden for tests
		generateUUID func() string
	}

	installIdentityOptions struct {
		replicas    uint
		trustDomain string

		issuanceLifetime   time.Duration
		clockSkewAllowance time.Duration

		trustPEMFile, crtPEMFile, keyPEMFile string
	}
)

const (
	prometheusImage                   = "prom/prometheus:v2.7.1"
	prometheusProxyOutboundCapacity   = 10000
	defaultControllerReplicas         = 1
	defaultHAControllerReplicas       = 3
	defaultIdentityTrustDomain        = "cluster.local"
	defaultIdentityIssuanceLifetime   = 24 * time.Hour
	defaultIdentityClockSkewAllowance = 20 * time.Second
)

// newInstallOptionsWithDefaults initializes install options with default
// control plane and proxy options.
//
// These options may be overridden on the CLI at install-time and will be
// persisted in Linkerd's control plane configuration to be used at
// injection-time.
// :custom
func newInstallOptionsWithDefaults(namespace string) *installOptions {
	return &installOptions{
		// :custom
		installNamespace:   namespace,
		controllerReplicas: defaultControllerReplicas,
		controllerLogLevel: "info",
		proxyAutoInject:    false,
		highAvailability:   false,
		controllerUID:      2103,
		disableH2Upgrade:   false,
		noInitContainer:    false,
		proxyConfigOptions: &proxyConfigOptions{
			// :custom:
			// set here because they use a runtime linked version
			linkerdVersion:     Version_stable230,
			ignoreCluster:      false,
			proxyImage:         defaultDockerRegistry + "/proxy",
			initImage:          defaultDockerRegistry + "/proxy-init",
			dockerRegistry:     defaultDockerRegistry,
			imagePullPolicy:    "IfNotPresent",
			ignoreInboundPorts: nil,
			// :custom:
			// ilackarms: added here because we are not using linkerd's rendering function but our own
			// Skip outbound port 443 to enable Kubernetes API access without the proxy.
			// Once Kubernetes supports sidecar containers, this may be removed, as that
			// will guarantee the proxy is running prior to control-plane startup.
			ignoreOutboundPorts:    []uint{443},
			proxyUID:               2102,
			proxyLogLevel:          "warn,linkerd2_proxy=info",
			proxyControlPort:       4190,
			proxyAdminPort:         4191,
			proxyInboundPort:       4143,
			proxyOutboundPort:      4140,
			proxyCPURequest:        "",
			proxyMemoryRequest:     "",
			proxyCPULimit:          "",
			proxyMemoryLimit:       "",
			enableExternalProfiles: false,
		},
		identityOptions: newInstallIdentityOptionsWithDefaults(),

		generateUUID: func() string {
			id, err := uuid.NewRandom()
			if err != nil {
				log.Fatalf("Could not generate UUID: %s", err)
			}
			return id.String()
		},
	}
}

func newInstallIdentityOptionsWithDefaults() *installIdentityOptions {
	return &installIdentityOptions{
		trustDomain:        defaultIdentityTrustDomain,
		issuanceLifetime:   defaultIdentityIssuanceLifetime,
		clockSkewAllowance: defaultIdentityClockSkewAllowance,
	}
}

func (options *installOptions) validateAndBuild() (*installValues, *pb.All, error) {
	if err := options.validate(); err != nil {
		return nil, nil, err
	}

	identityValues, err := options.identityOptions.validateAndBuild(options.installNamespace)
	if err != nil {
		return nil, nil, err
	}

	configs := options.configs(identityValues.toIdentityContext())

	values, err := options.buildValuesWithoutIdentity(configs)
	if err != nil {
		return nil, nil, err
	}
	values.Identity = identityValues

	return values, configs, nil
}

// installOnlyFlagSet includes flags that are only accessible at install-time
// and not at upgrade-time.
func (options *installOptions) installOnlyFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("install-only", pflag.ExitOnError)

	flags.StringVar(
		&options.identityOptions.trustDomain, "identity-trust-domain", options.identityOptions.trustDomain,
		"Configures the name suffix used for identities.",
	)
	flags.StringVar(
		&options.identityOptions.trustPEMFile, "identity-trust-anchors-file", options.identityOptions.trustPEMFile,
		"A path to a PEM-encoded file containing Linkerd Identity trust anchors (generated by default)",
	)
	flags.StringVar(
		&options.identityOptions.crtPEMFile, "identity-issuer-certificate-file", options.identityOptions.crtPEMFile,
		"A path to a PEM-encoded file containing the Linkerd Identity issuer certificate (generated by default)",
	)
	flags.StringVar(
		&options.identityOptions.keyPEMFile, "identity-issuer-key-file", options.identityOptions.keyPEMFile,
		"A path to a PEM-encoded file containing the Linkerd Identity issuer private key (generated by default)",
	)

	flags.BoolVar(
		&options.ignoreCluster, "ignore-cluster", options.ignoreCluster,
		"Ignore the current Kubernetes cluster when checking for existing cluster configuration (default false)",
	)

	return flags
}

func (options *installOptions) validate() error {
	if options.identityOptions == nil {
		// Programmer error: identityOptions may be empty, but it must be set by the constructor.
		panic("missing identity options")
	}

	if _, err := log.ParseLevel(options.controllerLogLevel); err != nil {
		return fmt.Errorf("--controller-log-level must be one of: panic, fatal, error, warn, info, debug")
	}

	if err := options.proxyConfigOptions.validate(); err != nil {
		return err
	}
	if options.proxyLogLevel == "" {
		return errors.Errorf("--proxy-log-level must not be empty")
	}

	if options.highAvailability {
		if options.controllerReplicas == defaultControllerReplicas {
			options.controllerReplicas = defaultHAControllerReplicas
		}

		if options.proxyCPURequest == "" {
			options.proxyCPURequest = "100m"
		}

		if options.proxyMemoryRequest == "" {
			options.proxyMemoryRequest = "20Mi"
		}
	}

	options.identityOptions.replicas = options.controllerReplicas

	return nil
}

func (options *installOptions) buildValuesWithoutIdentity(configs *pb.All) (*installValues, error) {
	globalJSON, proxyJSON, installJSON, err := config.ToJSON(configs)
	if err != nil {
		return nil, err
	}

	values := &installValues{
		// Container images:
		ControllerImage: fmt.Sprintf("%s/controller:%s", options.dockerRegistry, options.linkerdVersion),
		WebImage:        fmt.Sprintf("%s/web:%s", options.dockerRegistry, options.linkerdVersion),
		GrafanaImage:    fmt.Sprintf("%s/grafana:%s", options.dockerRegistry, options.linkerdVersion),
		PrometheusImage: prometheusImage,
		ImagePullPolicy: options.imagePullPolicy,

		// Kubernetes labels/annotations/resourcse:
		CreatedByAnnotation:      k8s.CreatedByAnnotation,
		CliVersion:               k8s.CreatedByAnnotationValue(),
		ControllerComponentLabel: k8s.ControllerComponentLabel,
		ProxyContainerName:       k8s.ProxyContainerName,
		ProxyInjectAnnotation:    k8s.ProxyInjectAnnotation,
		ProxyInjectDisabled:      k8s.ProxyInjectDisabled,

		// Controller configuration:
		// :custom
		Namespace:              options.installNamespace,
		UUID:                   configs.GetInstall().GetUuid(),
		ControllerReplicas:     options.controllerReplicas,
		ControllerLogLevel:     options.controllerLogLevel,
		ControllerUID:          options.controllerUID,
		EnableH2Upgrade:        !options.disableH2Upgrade,
		NoInitContainer:        options.noInitContainer,
		ProxyAutoInjectEnabled: options.proxyAutoInject,
		PrometheusLogLevel:     toPromLogLevel(options.controllerLogLevel),

		Configs: configJSONs{
			Global:  globalJSON,
			Proxy:   proxyJSON,
			Install: installJSON,
		},

		DestinationResources:   &resources{},
		GrafanaResources:       &resources{},
		IdentityResources:      &resources{},
		PrometheusResources:    &resources{},
		ProxyInjectorResources: &resources{},
		PublicAPIResources:     &resources{},
		SPValidatorResources:   &resources{},
		TapResources:           &resources{},
		WebResources:           &resources{},
	}

	if options.highAvailability {
		defaultConstraints := &resources{
			CPU:    constraints{Request: "100m"},
			Memory: constraints{Request: "50Mi"},
		}
		// Copy constraints to each so that further modification isn't global.
		*values.DestinationResources = *defaultConstraints
		*values.GrafanaResources = *defaultConstraints
		*values.ProxyInjectorResources = *defaultConstraints
		*values.PublicAPIResources = *defaultConstraints
		*values.SPValidatorResources = *defaultConstraints
		*values.TapResources = *defaultConstraints
		*values.WebResources = *defaultConstraints

		// The identity controller maintains no internal state, so it need not request
		// 50Mi.
		*values.IdentityResources = *defaultConstraints
		values.IdentityResources.Memory = constraints{Request: "10Mi"}

		values.PrometheusResources = &resources{
			CPU:    constraints{Request: "300m"},
			Memory: constraints{Request: "300Mi"},
		}
	}

	return values, nil
}

func toPromLogLevel(level string) string {
	switch level {
	case "panic", "fatal":
		return "error"
	default:
		return level
	}
}

func (options *installOptions) configs(identity *pb.IdentityContext) *pb.All {
	return &pb.All{
		Global:  options.globalConfig(identity),
		Proxy:   options.proxyConfig(),
		Install: options.installConfig(),
	}
}

func (options *installOptions) globalConfig(identity *pb.IdentityContext) *pb.Global {
	var autoInjectContext *pb.AutoInjectContext
	if options.proxyAutoInject {
		autoInjectContext = &pb.AutoInjectContext{}
	}

	return &pb.Global{
		LinkerdNamespace:  options.installNamespace,
		AutoInjectContext: autoInjectContext,
		CniEnabled:        options.noInitContainer,
		Version:           options.linkerdVersion,
		IdentityContext:   identity,
	}
}

func (options *installOptions) installConfig() *pb.Install {
	installID := ""
	if options.generateUUID != nil {
		installID = options.generateUUID()
	}

	return &pb.Install{
		Uuid:       installID,
		CliVersion: version.Version,
		Flags:      options.recordedFlags,
	}
}

func (options *installOptions) proxyConfig() *pb.Proxy {
	ignoreInboundPorts := []*pb.Port{}
	for _, port := range options.ignoreInboundPorts {
		ignoreInboundPorts = append(ignoreInboundPorts, &pb.Port{Port: uint32(port)})
	}

	ignoreOutboundPorts := []*pb.Port{}
	for _, port := range options.ignoreOutboundPorts {
		ignoreOutboundPorts = append(ignoreOutboundPorts, &pb.Port{Port: uint32(port)})
	}

	return &pb.Proxy{
		ProxyImage: &pb.Image{
			ImageName:  registryOverride(options.proxyImage, options.dockerRegistry),
			PullPolicy: options.imagePullPolicy,
		},
		ProxyInitImage: &pb.Image{
			ImageName:  registryOverride(options.initImage, options.dockerRegistry),
			PullPolicy: options.imagePullPolicy,
		},
		ControlPort: &pb.Port{
			Port: uint32(options.proxyControlPort),
		},
		IgnoreInboundPorts:  ignoreInboundPorts,
		IgnoreOutboundPorts: ignoreOutboundPorts,
		InboundPort: &pb.Port{
			Port: uint32(options.proxyInboundPort),
		},
		AdminPort: &pb.Port{
			Port: uint32(options.proxyAdminPort),
		},
		OutboundPort: &pb.Port{
			Port: uint32(options.proxyOutboundPort),
		},
		Resource: &pb.ResourceRequirements{
			RequestCpu:    options.proxyCPURequest,
			RequestMemory: options.proxyMemoryRequest,
			LimitCpu:      options.proxyCPULimit,
			LimitMemory:   options.proxyMemoryLimit,
		},
		ProxyUid: options.proxyUID,
		LogLevel: &pb.LogLevel{
			Level: options.proxyLogLevel,
		},
		DisableExternalProfiles: !options.enableExternalProfiles,
	}
}

func (idopts *installIdentityOptions) validate() error {
	if idopts == nil {
		return nil
	}

	if idopts.trustDomain != "" {
		if errs := validation.IsDNS1123Subdomain(idopts.trustDomain); len(errs) > 0 {
			return fmt.Errorf("invalid trust domain '%s': %s", idopts.trustDomain, errs[0])
		}
	}

	if idopts.trustPEMFile != "" || idopts.crtPEMFile != "" || idopts.keyPEMFile != "" {
		if idopts.trustPEMFile == "" {
			return errors.Errorf("a trust anchors file must be specified if other credentials are provided")
		}
		if idopts.crtPEMFile == "" {
			return errors.Errorf("a certificate file must be specified if other credentials are provided")
		}
		if idopts.keyPEMFile == "" {
			return errors.Errorf("a private key file must be specified if other credentials are provided")
		}

		for _, f := range []string{idopts.trustPEMFile, idopts.crtPEMFile, idopts.keyPEMFile} {
			stat, err := os.Stat(f)
			if err != nil {
				return fmt.Errorf("missing file: %s", err)
			}
			if stat.IsDir() {
				return fmt.Errorf("not a file: %s", f)
			}
		}
	}

	return nil
}

func (idopts *installIdentityOptions) validateAndBuild(installNamespace string) (*installIdentityValues, error) {
	if idopts == nil {
		return nil, nil
	}

	if err := idopts.validate(); err != nil {
		return nil, err
	}

	if idopts.trustPEMFile != "" && idopts.crtPEMFile != "" && idopts.keyPEMFile != "" {
		return idopts.readValues(installNamespace)
	}

	return idopts.genValues(installNamespace)
}

func (idopts *installIdentityOptions) issuerName(installNamespace string) string {
	return fmt.Sprintf("identity.%s.%s", installNamespace, idopts.trustDomain)
}

func (idopts *installIdentityOptions) genValues(installNamespace string) (*installIdentityValues, error) {
	root, err := tls.GenerateRootCAWithDefaults(idopts.issuerName(installNamespace))
	if err != nil {
		return nil, fmt.Errorf("failed to generate root certificate for identity: %s", err)
	}

	return &installIdentityValues{
		Replicas:        idopts.replicas,
		TrustDomain:     idopts.trustDomain,
		TrustAnchorsPEM: root.Cred.Crt.EncodeCertificatePEM(),
		Issuer: &issuerValues{
			ClockSkewAllowance:  idopts.clockSkewAllowance.String(),
			IssuanceLifetime:    idopts.issuanceLifetime.String(),
			CrtExpiryAnnotation: k8s.IdentityIssuerExpiryAnnotation,

			KeyPEM: root.Cred.EncodePrivateKeyPEM(),
			CrtPEM: root.Cred.Crt.EncodeCertificatePEM(),

			CrtExpiry: root.Cred.Crt.Certificate.NotAfter,
		},
	}, nil
}

// readValues attempts to read an issuer configuration from disk
// to produce an `installIdentityValues`.
//
// The identity options must have already been validated.
func (idopts *installIdentityOptions) readValues(installNamespace string) (*installIdentityValues, error) {
	creds, err := tls.ReadPEMCreds(idopts.keyPEMFile, idopts.crtPEMFile)
	if err != nil {
		return nil, err
	}

	trustb, err := ioutil.ReadFile(idopts.trustPEMFile)
	if err != nil {
		return nil, err
	}
	trustAnchorsPEM := string(trustb)
	roots, err := tls.DecodePEMCertPool(trustAnchorsPEM)
	if err != nil {
		return nil, err
	}

	if err := creds.Verify(roots, idopts.issuerName(installNamespace)); err != nil {
		return nil, fmt.Errorf("invalid credentials: %s", err)
	}

	return &installIdentityValues{
		Replicas:        idopts.replicas,
		TrustDomain:     idopts.trustDomain,
		TrustAnchorsPEM: trustAnchorsPEM,
		Issuer: &issuerValues{
			ClockSkewAllowance:  idopts.clockSkewAllowance.String(),
			IssuanceLifetime:    idopts.issuanceLifetime.String(),
			CrtExpiryAnnotation: k8s.IdentityIssuerExpiryAnnotation,

			KeyPEM: creds.EncodePrivateKeyPEM(),
			CrtPEM: creds.EncodeCertificatePEM(),

			CrtExpiry: creds.Crt.Certificate.NotAfter,
		},
	}, nil
}

func (idvals *installIdentityValues) toIdentityContext() *pb.IdentityContext {
	if idvals == nil {
		return nil
	}

	il, err := time.ParseDuration(idvals.Issuer.IssuanceLifetime)
	if err != nil {
		il = defaultIdentityIssuanceLifetime
	}

	csa, err := time.ParseDuration(idvals.Issuer.ClockSkewAllowance)
	if err != nil {
		csa = defaultIdentityClockSkewAllowance
	}

	return &pb.IdentityContext{
		TrustDomain:        idvals.TrustDomain,
		TrustAnchorsPem:    idvals.TrustAnchorsPEM,
		IssuanceLifetime:   ptypes.DurationProto(il),
		ClockSkewAllowance: ptypes.DurationProto(csa),
	}
}

const (
	defaultDockerRegistry = "gcr.io/linkerd-io"
)

var (
	kubeconfigPath            string
	kubeContext               string
	alphaNumDashDot           = regexp.MustCompile(`^[\.a-zA-Z0-9-]+$`)
	alphaNumDashDotSlashColon = regexp.MustCompile(`^[\./a-zA-Z0-9-:]+$`)

	// Full Rust log level syntax at
	// https://docs.rs/env_logger/0.6.0/env_logger/#enabling-logging
	r                  = strings.NewReplacer("\t", "", "\n", "")
	validProxyLogLevel = regexp.MustCompile(r.Replace(`
		^(
			(
				(trace|debug|warn|info|error)|
				(\w|::)+|
				((\w|::)+=(trace|debug|warn|info|error))
			)(?:,|$)
		)+$`))
)

// proxyConfigOptions holds values for command line flags that apply to both the
// install and inject commands. All fields in this struct should have
// corresponding flags added in the addProxyConfigFlags func later in this file.
type proxyConfigOptions struct {
	linkerdVersion         string
	proxyImage             string
	initImage              string
	dockerRegistry         string
	imagePullPolicy        string
	ignoreInboundPorts     []uint
	ignoreOutboundPorts    []uint
	proxyUID               int64
	proxyLogLevel          string
	proxyInboundPort       uint
	proxyOutboundPort      uint
	proxyControlPort       uint
	proxyAdminPort         uint
	proxyCPURequest        string
	proxyMemoryRequest     string
	proxyCPULimit          string
	proxyMemoryLimit       string
	enableExternalProfiles bool
	// ignoreCluster is not validated by validate().
	ignoreCluster bool
}

func (options *proxyConfigOptions) validate() error {
	if options.linkerdVersion != "" && !alphaNumDashDot.MatchString(options.linkerdVersion) {
		return fmt.Errorf("%s is not a valid version", options.linkerdVersion)
	}

	if options.dockerRegistry != "" && !alphaNumDashDotSlashColon.MatchString(options.dockerRegistry) {
		return fmt.Errorf("%s is not a valid Docker registry. The url can contain only letters, numbers, dash, dot, slash and colon", options.dockerRegistry)
	}

	if options.imagePullPolicy != "" && options.imagePullPolicy != "Always" && options.imagePullPolicy != "IfNotPresent" && options.imagePullPolicy != "Never" {
		return fmt.Errorf("--image-pull-policy must be one of: Always, IfNotPresent, Never")
	}

	if options.proxyCPURequest != "" {
		if _, err := k8sResource.ParseQuantity(options.proxyCPURequest); err != nil {
			return fmt.Errorf("Invalid cpu request '%s' for --proxy-cpu-request flag", options.proxyCPURequest)
		}
	}

	if options.proxyMemoryRequest != "" {
		if _, err := k8sResource.ParseQuantity(options.proxyMemoryRequest); err != nil {
			return fmt.Errorf("Invalid memory request '%s' for --proxy-memory-request flag", options.proxyMemoryRequest)
		}
	}

	if options.proxyCPULimit != "" {
		cpuLimit, err := k8sResource.ParseQuantity(options.proxyCPULimit)
		if err != nil {
			return fmt.Errorf("Invalid cpu limit '%s' for --proxy-cpu-limit flag", options.proxyCPULimit)
		}
		if options.proxyCPURequest != "" {
			// Not checking for error because option proxyCPURequest was already validated
			if cpuRequest, _ := k8sResource.ParseQuantity(options.proxyCPURequest); cpuRequest.MilliValue() > cpuLimit.MilliValue() {
				return fmt.Errorf("The cpu limit '%s' cannot be lower than the cpu request '%s'", options.proxyCPULimit, options.proxyCPURequest)
			}
		}
	}

	if options.proxyMemoryLimit != "" {
		memoryLimit, err := k8sResource.ParseQuantity(options.proxyMemoryLimit)
		if err != nil {
			return fmt.Errorf("Invalid memory limit '%s' for --proxy-memory-limit flag", options.proxyMemoryLimit)
		}
		if options.proxyMemoryRequest != "" {
			// Not checking for error because option proxyMemoryRequest was already validated
			if memoryRequest, _ := k8sResource.ParseQuantity(options.proxyMemoryRequest); memoryRequest.Value() > memoryLimit.Value() {
				return fmt.Errorf("The memory limit '%s' cannot be lower than the memory request '%s'", options.proxyMemoryLimit, options.proxyMemoryRequest)
			}
		}
	}

	if options.proxyLogLevel != "" && !validProxyLogLevel.MatchString(options.proxyLogLevel) {
		return fmt.Errorf("\"%s\" is not a valid proxy log level - for allowed syntax check https://docs.rs/env_logger/0.6.0/env_logger/#enabling-logging",
			options.proxyLogLevel)
	}

	return nil
}

// registryOverride replaces the registry of the provided image if the image is
// using the default registry and the provided registry is not the default.
func registryOverride(image, registry string) string {
	return strings.Replace(image, defaultDockerRegistry, registry, 1)
}

// used for injection
type injector struct {
	configs               *pb.All
	overrideAnnotations   map[string]string
	proxyOutboundCapacity map[string]uint
}

func (rt injector) inject(bytes []byte) ([]byte, error) {
	conf := inject.NewResourceConfig(rt.configs, inject.OriginCLI)
	if len(rt.proxyOutboundCapacity) > 0 {
		conf = conf.WithProxyOutboundCapacity(rt.proxyOutboundCapacity)
	}

	report, err := conf.ParseMetaAndYAML(bytes)
	if err != nil {
		return nil, err
	}

	if !report.Injectable() {
		return bytes, nil
	}

	conf.AppendPodAnnotations(map[string]string{
		k8s.CreatedByAnnotation: k8s.CreatedByAnnotationValue(),
	})
	if len(rt.overrideAnnotations) > 0 {
		conf.AppendPodAnnotations(rt.overrideAnnotations)
	}

	p, err := conf.GetPatch(bytes)
	if err != nil {
		return nil, err
	}
	if p.IsEmpty() {
		return bytes, nil
	}

	patchJSON, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	if patchJSON == nil {
		return bytes, nil
	}
	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		return nil, err
	}
	origJSON, err := yaml.YAMLToJSON(bytes)
	if err != nil {
		return nil, err
	}
	injectedJSON, err := patch.Apply(origJSON)
	if err != nil {
		return nil, err
	}
	injectedYAML, err := conf.JSONToYAML(injectedJSON)
	if err != nil {
		return nil, err
	}
	return injectedYAML, nil
}

// copied from github.com/linkerd/linkerd2
// processYAML takes an input stream of YAML, outputting injected/uninjected YAML to out.
func processYAML(in io.Reader, out io.Writer, rt injector) error {
	reader := yamlDecoder.NewYAMLReader(bufio.NewReaderSize(in, 4096))

	// Iterate over all YAML objects in the input
	for {
		// Read a single YAML object
		bytes, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var result []byte

		isList, err := kindIsList(bytes)
		if err != nil {
			return err
		}
		if isList {
			result, err = processList(bytes, rt)
		} else {
			result, err = rt.inject(bytes)
		}
		if err != nil {
			return err
		}
		out.Write(result)
		out.Write([]byte("---\n"))
	}

	return nil
}

func kindIsList(bytes []byte) (bool, error) {
	var meta metav1.TypeMeta
	if err := yaml.Unmarshal(bytes, &meta); err != nil {
		return false, err
	}
	return meta.Kind == "List", nil
}

func processList(bytes []byte, rt injector) ([]byte, error) {
	var sourceList corev1.List
	if err := yaml.Unmarshal(bytes, &sourceList); err != nil {
		return nil, err
	}

	items := []runtime.RawExtension{}

	for _, item := range sourceList.Items {
		result, err := rt.inject(item.Raw)
		if err != nil {
			return nil, err
		}

		// At this point, we have yaml. The kubernetes internal representation is
		// json. Because we're building a list from RawExtensions, the yaml needs
		// to be converted to json.
		injected, err := yaml.YAMLToJSON(result)
		if err != nil {
			return nil, err
		}

		items = append(items, runtime.RawExtension{Raw: injected})
	}

	sourceList.Items = items
	result, err := yaml.Marshal(sourceList)
	if err != nil {
		return nil, err
	}
	return result, nil
}
