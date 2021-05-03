package gloo_mesh

import (
	"context"
	"fmt"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/uninstall"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	gloo_context "github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"io"
	"istio.io/istio/pkg/test/framework/components/cluster"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/util/retry"
	v1 "k8s.io/api/core/v1"
	kubeApiMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var (
	_ gloo_context.GlooMeshInstance = &glooMeshInstance{}
	_ io.Closer                     = &glooMeshInstance{}
)

const namespace = "gloo-mesh"

type glooMeshInstance struct {
	id             resource.ID
	instanceConfig InstanceConfig
}
type InstanceConfig struct {
	managementPlane                   bool
	controlPlaneKubeConfigPath        string
	managementPlaneKubeConfigPath     string
	version                           string
	clusterDomain                     string
	managementPlaneRelayServerAddress string
	cluster                           cluster.Cluster
}

func newInstance(ctx resource.Context, instanceConfig InstanceConfig, licenceKey string) (gloo_context.GlooMeshInstance, error) {
	var err error
	i := &glooMeshInstance{}
	i.id = ctx.TrackResource(i)
	i.instanceConfig = instanceConfig
	if i.instanceConfig.managementPlane {
		// deploy enterprise version
		if err := i.deployManagementPlane(licenceKey); err != nil {
			return nil, err
		}
	} else {
		if err := i.deployControlPlane(licenceKey); err != nil {
			return nil, err
		}
	}
	return i, err
}

func (i *glooMeshInstance) deployManagementPlane(licenceKey string) error {
	options := &install.Options{
		GlobalFlags:     &utils.GlobalFlags{Verbose: true},
		DryRun:          false,
		KubeCfgPath:     i.instanceConfig.managementPlaneKubeConfigPath,
		KubeContext:     "",
		Namespace:       namespace,
		ChartPath:       "",
		ChartValuesFile: "",
		Version:         i.instanceConfig.version,
		ReleaseName:     fmt.Sprintf("%s-mp", i.instanceConfig.cluster.Name()),
		AgentChartPath:  "",
		AgentValuesPath: "",
		Register:        false,
		ClusterName:     "",
		ClusterDomain:   "",
	}

	if licenceKey != "" {
		if err := install.InstallEnterprise(context.Background(), install.EnterpriseOptions{
			Options:            options,
			LicenseKey:         licenceKey,
			SkipUI:             false,
			SkipRBAC:           false,
			RelayServerAddress: "",
		}); err != nil {
			return err
		}
	} else {
		if err := install.InstallCommunity(context.Background(), install.CommunityOptions{
			Options:            options,
			AgentCrdsChartPath: "",
			ApiServerAddress:   "",
		}); err != nil {
			return err
		}
	}

	return nil
}

func (i *glooMeshInstance) deployControlPlane(licenceKey string) error {

	options := registration.Options{
		KubeConfigPath:         i.instanceConfig.controlPlaneKubeConfigPath,
		MgmtKubeConfigPath:     i.instanceConfig.managementPlaneKubeConfigPath,
		MgmtContext:            "",
		MgmtNamespace:          namespace,
		RemoteContext:          "",
		RemoteNamespace:        namespace,
		Version:                "1.0.5",
		AgentCrdsChartPath:     "",
		AgentChartPathOverride: "",
		AgentChartValuesPath:   "",
		ApiServerAddress:       "",
		ClusterName:            i.instanceConfig.cluster.Name(),
		ClusterDomain:          "",
		Verbose:                false,
	}

	if licenceKey != "" {
		if err := enterprise.RegisterCluster(context.Background(), enterprise.RegistrationOptions{
			Options:                   options,
			AgentChartPathOverride:    "",
			AgentChartValuesPath:      "",
			RelayServerAddress:        i.instanceConfig.managementPlaneRelayServerAddress,
			RelayServerInsecure:       false,
			RootCASecretName:          "",
			RootCASecretNamespace:     "",
			ClientCertSecretName:      "",
			ClientCertSecretNamespace: "",
			TokenSecretName:           "",
			TokenSecretNamespace:      "",
			TokenSecretKey:            "",
			ReleaseName:               fmt.Sprintf("%s-cp", i.instanceConfig.cluster.Name()),
		}); err != nil {
			return err
		}

		// TODO this isnt needed but we should do some sort of sanity check that MP came up correctly
		if err := i.waitForSecretsForNamespace([]string{
			"rbac-webhook",
			"relay-server-tls-secret",
			"relay-tls-signing-secret",
			"relay-identity-token-secret",
			"relay-client-tls-secret",
			"gloo-mesh-enterprise-license",
			"relay-root-tls-secret",
		}, namespace); err != nil {
			return err
		}
	} else {

		registrant, err := registration.NewRegistrant(options)
		if err != nil {
			return err
		}

		return registrant.RegisterCluster(context.Background())
	}

	return nil
}

func (i *glooMeshInstance) ID() resource.ID {
	return i.id
}

func (i *glooMeshInstance) GetRelayServerAddress() (string, error) {
	if !i.instanceConfig.managementPlane {
		return "", fmt.Errorf("cluster does not have a management plane")
	}
	svcName := "enterprise-networking"

	svc, err := i.instanceConfig.cluster.CoreV1().Services(namespace).Get(context.TODO(), svcName, kubeApiMeta.GetOptions{})
	if err != nil {
		return "", err
	}

	// This probably wont work in all situations
	return serviceIngressToAddress(svc)
}

func (i *glooMeshInstance) IsManagementPlane() bool {
	return i.instanceConfig.managementPlane
}

func (i *glooMeshInstance) GetKubeConfig() string {
	return i.instanceConfig.managementPlaneKubeConfigPath
}

// Close implements io.Closer.
func (i *glooMeshInstance) Close() error {

	if i.instanceConfig.managementPlane {
		kubeConfig := i.instanceConfig.managementPlaneKubeConfigPath
		releaseName := fmt.Sprintf("%s-mp", i.instanceConfig.cluster.Name())
		return uninstall.Uninstall(context.Background(), &uninstall.Options{
			Verbose:     true,
			KubeCfgPath: kubeConfig,
			KubeContext: "",
			Namespace:   "gloo-mesh",
			ReleaseName: releaseName,
		})
	} else {
		releaseName := fmt.Sprintf("%s-cp", i.instanceConfig.cluster.Name())
		enterprise.DeregisterCluster(context.Background(), enterprise.RegistrationOptions{
			Options: registration.Options{
				KubeConfigPath:         i.instanceConfig.controlPlaneKubeConfigPath,
				MgmtKubeConfigPath:     i.instanceConfig.managementPlaneKubeConfigPath,
				MgmtContext:            "",
				MgmtNamespace:          namespace,
				RemoteContext:          "",
				RemoteNamespace:        namespace,
				Version:                "",
				AgentCrdsChartPath:     "",
				AgentChartPathOverride: "",
				AgentChartValuesPath:   "",
				ApiServerAddress:       "",
				ClusterName:            i.instanceConfig.cluster.Name(),
				ClusterDomain:          "",
				Verbose:                false,
			},
			AgentChartPathOverride:    "",
			AgentChartValuesPath:      "",
			RelayServerAddress:        i.instanceConfig.managementPlaneRelayServerAddress,
			RelayServerInsecure:       false,
			RootCASecretName:          "",
			RootCASecretNamespace:     "",
			ClientCertSecretName:      "",
			ClientCertSecretNamespace: "",
			TokenSecretName:           "",
			TokenSecretNamespace:      "",
			TokenSecretKey:            "",
			ReleaseName:               releaseName,
		})
	}

	return nil
}

func serviceIngressToAddress(svc *v1.Service) (string, error) {

	port := "9900"
	var address string
	ingress := svc.Status.LoadBalancer.Ingress
	if len(ingress) == 0 {
		return "", fmt.Errorf("no loadBalancer.ingress status reported for service")
	}

	// If the Ip address is set in the ingress, use that
	if ingress[0].IP != "" {
		address = ingress[0].IP
	} else {
		// Otherwise use the hostname
		address = ingress[0].Hostname
	}
	return fmt.Sprintf("%s:%s", address, port), nil
}

// wait until secrets are created before returning
func (i *glooMeshInstance) waitForSecretsForNamespace(secrets []string, ns string) error {
	for _, s := range secrets {
		if err := retry.UntilSuccess(func() error {

			_, err := i.instanceConfig.cluster.CoreV1().Secrets(ns).Get(context.TODO(), s, kubeApiMeta.GetOptions{})
			if err == nil {
				return nil
			}

			return nil
		}, retry.Timeout(time.Minute)); err != nil {
			return fmt.Errorf("failed to find secret %s %s", s, err.Error())
		}
	}
	return nil
}
