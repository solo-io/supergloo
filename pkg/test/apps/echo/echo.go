package echo

import (
	"embed"
	"strconv"

	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"github.com/solo-io/gloo-mesh/pkg/test/tlssecret"
	"istio.io/istio/pkg/config/protocol"
	common2 "istio.io/istio/pkg/test/echo/common"
	"istio.io/istio/pkg/test/framework/components/cluster"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/components/echo/echoboot"
	"istio.io/istio/pkg/test/framework/components/namespace"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/scopes"
)

var (
	//go:embed certs
	certFiles embed.FS
)

func DeployEchos(deploymentCtx *context.DeploymentContext) resource.SetupFn {
	return func(ctx resource.Context) error {
		if deploymentCtx == nil {
			*deploymentCtx = context.DeploymentContext{}
		}
		echoCtx := &context.EchoDeploymentContext{}

		if err := createNamespaces(ctx, echoCtx); err != nil {
			return err
		}
		if err := generateTLSCertificates(ctx, "echo-certs", echoCtx.AppNamespace); err != nil {
			return err
		}
		if err := generateTLSCertificates(ctx, "echo-certs", echoCtx.SubsetNamespace); err != nil {
			return err
		}

		apps, err := deployApplications(ctx, echoCtx)
		if err != nil {
			return err
		}

		echoCtx.Deployments = apps
		deploymentCtx.EchoContext = echoCtx
		return nil
	}
}

func deployApplications(ctx resource.Context, echoCtx *context.EchoDeploymentContext) (echo.Instances, error) {
	builder := echoboot.NewBuilder(ctx)
	for _, c := range ctx.Clusters() {
		frontendApp := newEchoConfig("frontend", echoCtx.AppNamespace, c, true, false)
		if _, err := builder.With(nil, frontendApp).Build(); err != nil {
			scopes.Framework.Errorf("error setting up frontend echo %v", err.Error())
			return nil, err
		}

		backendApp := newEchoConfig("backend", echoCtx.AppNamespace, c, true, false)
		if _, err := builder.With(nil, backendApp).Build(); err != nil {
			scopes.Framework.Errorf("error setting up backend echo %v", err.Error())
			return nil, err
		}

		subsetApp := newEchoConfig("subset", echoCtx.SubsetNamespace, c, true, true)
		if _, err := builder.With(nil, subsetApp).Build(); err != nil {
			scopes.Framework.Errorf("error setting up subset echos %v", err.Error())
			return nil, err
		}

		nonMeshApp := newEchoConfig("no-mesh", echoCtx.NoMeshNamespace, c, false, false)
		if _, err := builder.With(nil, nonMeshApp).Build(); err != nil {
			scopes.Framework.Errorf("error setting up no mesh echo %v", err.Error())
			return nil, err
		}
	}

	apps, err := builder.Build()
	if err != nil {
		scopes.Framework.Errorf("error setting up echo apps %v", err.Error())
		return nil, err
	}
	return apps, nil
}

func createNamespaces(ctx resource.Context, echoCtx *context.EchoDeploymentContext) error {
	var err error

	if echoCtx.AppNamespace, err = namespace.New(ctx, namespace.Config{Prefix: "app", Inject: true}); err != nil {
		return err
	}

	if echoCtx.SubsetNamespace, err = namespace.New(ctx, namespace.Config{Prefix: "subset", Inject: true}); err != nil {
		return err
	}

	if echoCtx.NoMeshNamespace, err = namespace.New(ctx, namespace.Config{Prefix: "no-mesh", Inject: false}); err != nil {
		return err
	}

	return nil
}

func generateTLSCertificates(ctx resource.Context, secretName string, ns namespace.Instance) error {
	echoCrt, err := certFiles.ReadFile("certs/echo.crt")
	if err != nil {
		scopes.Framework.Error(err)
	}
	echoKey, err := certFiles.ReadFile("certs/echo.key")
	if err != nil {
		scopes.Framework.Error(err)
	}
	echoCA, err := certFiles.ReadFile("certs/echo-ca.crt")
	if err != nil {
		scopes.Framework.Error(err)
	}
	for _, c := range ctx.Clusters() {
		_, err = tlssecret.New(ctx, &tlssecret.Config{
			Namespace: ns.Name(),
			Name:      secretName,
			CACrt:     string(echoCA),
			TLSKey:    string(echoKey),
			TLSCert:   string(echoCrt),
			Cluster:   c,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func newEchoConfig(service string, ns namespace.Instance, cluster cluster.Cluster, hasSidecar bool, useSubsets bool) echo.Config {
	echoCrt, err := certFiles.ReadFile("certs/echo.crt")
	if err != nil {
		scopes.Framework.Error(err)
	}
	echoKey, err := certFiles.ReadFile("certs/echo.key")
	if err != nil {
		scopes.Framework.Error(err)
	}
	echoCA, err := certFiles.ReadFile("certs/echo-ca.crt")
	if err != nil {
		scopes.Framework.Error(err)
	}

	tlsSettings := &common2.TLSSettings{
		RootCert:   string(echoCA),
		ClientCert: string(echoCrt),
		Key:        string(echoKey),
	}
	var subset []echo.SubsetConfig
	if !hasSidecar {
		subset = []echo.SubsetConfig{
			{
				Annotations: map[echo.Annotation]*echo.AnnotationValue{
					echo.SidecarInject: {
						Value: strconv.FormatBool(false)},
				},
			},
		}
	} else if useSubsets {
		subset = []echo.SubsetConfig{
			{
				Version: "v1",
			},
			{
				Version: "v2",
			},
		}
	}

	return echo.Config{
		Namespace: ns,
		Service:   service,
		Ports: []echo.Port{
			{
				Name:     "http",
				Protocol: protocol.HTTP,
				// We use a port > 1024 to not require root
				ServicePort:  8090,
				InstancePort: 8090,
			},
			{
				// HTTPS port
				Name:         "https",
				Protocol:     protocol.HTTPS,
				ServicePort:  9443,
				InstancePort: 9443,
				TLS:          true,
			},
			{
				// TCP port
				Name:         "tcp",
				Protocol:     protocol.TCP,
				ServicePort:  9000,
				InstancePort: 9000,
				TLS:          false,
			},
		},
		Subsets:     subset,
		TLSSettings: tlsSettings,
		Cluster:     cluster,
	}
}
