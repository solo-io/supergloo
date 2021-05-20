package gloo_mesh

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/uninstall"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	gloo_context "github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/framework/components/cluster"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/util/retry"
	"istio.io/pkg/log"
	v1 "k8s.io/api/core/v1"
	kubeApiMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func newInstance(ctx resource.Context, instanceConfig InstanceConfig, licenseKey string) (gloo_context.GlooMeshInstance, error) {
	var err error
	i := &glooMeshInstance{}
	i.id = ctx.TrackResource(i)
	i.instanceConfig = instanceConfig
	if i.instanceConfig.managementPlane {
		// deploy enterprise version
		if err := i.deployManagementPlane(licenseKey); err != nil {
			return nil, err
		}
	} else {
		if err := i.deployControlPlane(licenseKey); err != nil {
			return nil, err
		}
	}
	return i, err
}

func (i *glooMeshInstance) deployManagementPlane(licenseKey string) error {
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

	if licenseKey != "" {
		if err := install.InstallEnterprise(context.Background(), install.EnterpriseOptions{
			Options:            options,
			LicenseKey:         licenseKey,
			SkipUI:             false,
			IncludeRBAC:        false,
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

func (i *glooMeshInstance) deployControlPlane(licenseKey string) error {

	options := registration.Options{
		KubeConfigPath:         i.instanceConfig.controlPlaneKubeConfigPath,
		MgmtKubeConfigPath:     i.instanceConfig.managementPlaneKubeConfigPath,
		MgmtContext:            "",
		MgmtNamespace:          namespace,
		RemoteContext:          "",
		RemoteNamespace:        namespace,
		Version:                i.instanceConfig.version,
		AgentCrdsChartPath:     "",
		AgentChartPathOverride: "",
		AgentChartValuesPath:   "",
		ApiServerAddress:       "",
		ClusterName:            i.instanceConfig.cluster.Name(),
		ClusterDomain:          "",
		Verbose:                false,
	}

	if licenseKey != "" {
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
func (i *glooMeshInstance) GetCluster() cluster.Cluster {
	return i.instanceConfig.cluster
}

// Close implements io.Closer.
func (i *glooMeshInstance) Close() error {
	// TODO need to clean up Solo CRDs
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
		if err := enterprise.DeregisterCluster(context.Background(), enterprise.RegistrationOptions{
			Options: registration.Options{
				KubeConfigPath:         i.instanceConfig.controlPlaneKubeConfigPath,
				MgmtKubeConfigPath:     i.instanceConfig.managementPlaneKubeConfigPath,
				MgmtContext:            "",
				MgmtNamespace:          namespace,
				RemoteContext:          "",
				RemoteNamespace:        namespace,
				Version:                i.instanceConfig.version,
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
		}); err != nil {
			log.Error(err)
		}
	}

	return nil
}

func serviceIngressToAddress(svc *v1.Service) (string, error) {

	port := "9900"
	var address string
	ingress := svc.Status.LoadBalancer.Ingress
	if len(ingress) == 0 {
		// Check for user-set external IPs
		externalIPs := svc.Spec.ExternalIPs
		if len(externalIPs) != 0 {
			address = svc.Spec.ExternalIPs[0]
		} else {
			return "", fmt.Errorf("no loadBalancer.ingress status reported for service. Please set an external IP on the service as a user if you are using a non-kubernetes load balancer.")
		}
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
