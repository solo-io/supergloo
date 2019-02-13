package install

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/supergloo/pkg2/kube"

	"github.com/solo-io/supergloo/pkg2/secret"

	"github.com/solo-io/solo-kit/pkg/utils/contextutils"

	"github.com/solo-io/supergloo/pkg2/install/linkerd2"

	"github.com/solo-io/supergloo/pkg2/install/istio"

	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg2/api/v1"
	"github.com/solo-io/supergloo/pkg2/install/consul"
	"github.com/solo-io/supergloo/pkg2/install/helm"

	istiov1 "github.com/solo-io/supergloo/pkg2/api/external/istio/encryption/v1"
	kube_client "github.com/solo-io/supergloo/pkg2/kube"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

type InstallSyncer struct {
	MeshClient        v1.MeshClient
	IstioSecretClient istiov1.IstioCacertsSecretClient
	RbacClient        kube.RbacClient
	NamespaceClient   kube.NamespaceClient
	HelmClient        helm.HelmClient

	istioInstaller    *istio.IstioInstaller
	linkerd2Installer *linkerd2.Linkerd2Installer
	consulInstaller   *consul.ConsulInstaller
}

func NewInstallSyncer(
	meshClient v1.MeshClient,
	istioSecretClient istiov1.IstioCacertsSecretClient,
	secretSyncer secret.SecretSyncer,
	rbacClient kube.RbacClient,
	namespaceClient kube.NamespaceClient,
	crdClient kube.CrdClient,
	helmClient helm.HelmClient) (*InstallSyncer, error) {
	istio, err := istio.NewIstioInstaller(crdClient, nil, secretSyncer)
	if err != nil {
		return nil, errors.Wrap(err, "setting up istio installer")
	}
	consul := &consul.ConsulInstaller{}
	linkerd2 := &linkerd2.Linkerd2Installer{}
	syncer := &InstallSyncer{
		MeshClient:        meshClient,
		IstioSecretClient: istioSecretClient,
		RbacClient:        rbacClient,
		NamespaceClient:   namespaceClient,
		HelmClient:        helmClient,

		istioInstaller:    istio,
		linkerd2Installer: linkerd2,
		consulInstaller:   consul,
	}
	return syncer, nil

}

func NewKubeInstallSyncer(meshClient v1.MeshClient, istioSecretClient istiov1.IstioCacertsSecretClient, kube kubernetes.Interface, apiExts apiexts.Interface) (*InstallSyncer, error) {
	crdClient := kube_client.NewKubeCrdClient(apiExts)
	rbacClient := kube_client.NewKubeRbacClient(kube)
	namespaceClient := kube_client.NewKubeNamespaceClient(kube)
	secretClient := kube_client.NewKubeSecretClient(kube)
	podClient := kube_client.NewKubePodClient(kube)

	secretSyncer := &secret.KubeSecretSyncer{
		SecretClient:      secretClient,
		PodClient:         podClient,
		IstioSecretClient: istioSecretClient,
	}

	helmClient := &helm.KubeHelmClient{}

	return NewInstallSyncer(meshClient, istioSecretClient, secretSyncer, rbacClient, namespaceClient, crdClient, helmClient)
}

type MeshInstaller interface {
	GetDefaultNamespace() string
	GetCrbName() string
	GetOverridesYaml(install *v1.Install) string
	DoPreHelmInstall(ctx context.Context, installNamespace string, install *v1.Install, secretList istiov1.IstioCacertsSecretList) error
}

func (syncer *InstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	secretList := snap.Istiocerts.List()
	ctx = contextutils.WithLogger(ctx, "install-syncer")
	for _, install := range snap.Installs.List() {
		err := syncer.syncInstall(ctx, install, secretList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (syncer *InstallSyncer) syncInstall(ctx context.Context, install *v1.Install, secretList istiov1.IstioCacertsSecretList) error {
	var meshInstaller MeshInstaller
	switch install.MeshType.(type) {
	case *v1.Install_Consul:
		meshInstaller = syncer.consulInstaller
	case *v1.Install_Istio:
		meshInstaller = syncer.istioInstaller
	case *v1.Install_Linkerd2:
		meshInstaller = syncer.linkerd2Installer
	default:
		return errors.Errorf("Unsupported mesh type %v", install.MeshType)
	}

	installEnabled := install.Enabled == nil || install.Enabled.Value

	mesh, meshErr := syncer.MeshClient.Read(install.Metadata.Namespace, install.Metadata.Name, clients.ReadOpts{Ctx: ctx})
	switch {
	case meshErr == nil && !installEnabled:
		if err := syncer.uninstallHelmRelease(ctx, mesh, install, meshInstaller); err != nil {
			return err
		}
		return syncer.MeshClient.Delete(mesh.Metadata.Namespace, mesh.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
	case meshErr == nil && installEnabled:
		if err := syncer.updateHelmRelease(ctx, install.ChartLocator, mesh.Metadata.Name, meshInstaller.GetOverridesYaml(install)); err != nil {
			return err
		}
	case meshErr != nil && installEnabled:
		releaseName, err := syncer.installHelmRelease(ctx, install, meshInstaller, secretList)
		if err != nil {
			return err
		}
		return syncer.createMesh(ctx, install, releaseName)
	}
	return nil
}

func (syncer *InstallSyncer) installHelmRelease(ctx context.Context, install *v1.Install, installer MeshInstaller, secretList istiov1.IstioCacertsSecretList) (string, error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("setting up namespace")
	// 1. Setup namespace
	installNamespace, err := syncer.setupInstallNamespace(install, installer)
	if err != nil {
		return "", err
	}

	// 2. Set up ClusterRoleBinding for that namespace
	// This is not cleaned up when deleting namespace so it may already exist on the system, don't fail
	crbName := installer.GetCrbName()
	if crbName != "" {
		err = syncer.RbacClient.CreateCrbIfNotExist(crbName, installNamespace)
		if err != nil {
			return "", errors.Wrap(err, "Error creating CRB")
		}
	}

	logger.Infof("helm pre-install")
	// 3. Do any pre-helm tasks
	err = installer.DoPreHelmInstall(ctx, installNamespace, install, secretList)
	if err != nil {
		return "", errors.Wrap(err, "Error doing pre-helm install steps")
	}

	logger.Infof("helm install")
	// 4. Install mesh via helm chart
	release, err := syncer.helmInstall(ctx, install.ChartLocator, install.Metadata.Name, installNamespace, installer.GetOverridesYaml(install))
	if err != nil {
		return "", errors.Wrap(err, "installing helm chart")
	}

	logger.Infof("finished installing %v", release)
	// 5. Do any additional steps
	return release, nil
}

func (syncer *InstallSyncer) setupInstallNamespace(install *v1.Install, installer MeshInstaller) (string, error) {
	installNamespace := getInstallNamespace(install, installer.GetDefaultNamespace())
	err := syncer.NamespaceClient.CreateNamespaceIfNotExist(installNamespace)
	if err != nil {
		return installNamespace, errors.Wrap(err, "Error setting up namespace")
	}
	return installNamespace, nil
}

func getInstallNamespace(install *v1.Install, defaultNamespace string) string {
	installNamespace := getInstallationNamespace(install)
	if installNamespace != "" {
		return installNamespace
	}
	return defaultNamespace
}

func getInstallationNamespace(install *v1.Install) (installationNamespace string) {
	switch x := install.MeshType.(type) {
	case *v1.Install_Istio:
		return x.Istio.InstallationNamespace
	case *v1.Install_Consul:
		return x.Consul.InstallationNamespace
	case *v1.Install_Linkerd2:
		return x.Linkerd2.InstallationNamespace
	default:
		//should never happen
		return ""
	}
}

func (syncer *InstallSyncer) CreateCrbIfNotExist(crbName string, namespaceName string) error {
	return syncer.RbacClient.CreateCrbIfNotExist(crbName, namespaceName)
}

func (syncer *InstallSyncer) helmInstall(ctx context.Context, chartLocator *v1.HelmChartLocator, releaseName string, installNamespace string, overridesYaml string) (string, error) {
	if chartLocator.GetChartPath() != nil {
		return syncer.HelmClient.InstallHelmRelease(ctx, chartLocator.GetChartPath().Path, releaseName, installNamespace, overridesYaml)
	}

	return "", errors.Errorf("Unsupported kind of chart locator")
}

func (syncer *InstallSyncer) updateHelmRelease(ctx context.Context, chartLocator *v1.HelmChartLocator, releaseName string, overridesYaml string) error {
	if chartLocator.GetChartPath() != nil {
		return syncer.HelmClient.UpdateHelmRelease(ctx, chartLocator.GetChartPath().Path, releaseName, overridesYaml)
	}
	return errors.Errorf("Unsupported kind of chart locator")
}

func (syncer *InstallSyncer) createMesh(ctx context.Context, install *v1.Install, releaseName string) error {
	mesh, err := GetMeshObject(install, releaseName)
	if err != nil {
		return err
	}
	_, err = syncer.MeshClient.Write(mesh, clients.WriteOpts{Ctx: ctx})
	return err
}

func GetMeshObject(install *v1.Install, releaseName string) (*v1.Mesh, error) {
	mesh := &v1.Mesh{
		Metadata: core.Metadata{
			Name:        install.Metadata.Name,
			Namespace:   install.Metadata.Namespace,
			Annotations: map[string]string{helm.ReleaseNameKey: releaseName},
		},
		Encryption: install.Encryption,
	}
	var err error
	switch x := install.MeshType.(type) {
	case *v1.Install_Istio:
		mesh.MeshType = &v1.Mesh_Istio{
			Istio: x.Istio,
		}
	case *v1.Install_Consul:
		mesh.MeshType = &v1.Mesh_Consul{
			Consul: x.Consul,
		}
	case *v1.Install_Linkerd2:
		mesh.MeshType = &v1.Mesh_Linkerd2{
			Linkerd2: x.Linkerd2,
		}
	default:
		err = errors.Errorf("Unsupported mesh type.")
	}
	return mesh, err
}

func (syncer *InstallSyncer) uninstallHelmRelease(ctx context.Context, mesh *v1.Mesh, install *v1.Install, meshInstaller MeshInstaller) error {
	releaseName := mesh.Metadata.Annotations[helm.ReleaseNameKey]
	syncer.HelmClient.DeleteHelmRelease(ctx, releaseName)
	// Install may be into ns that can't be deleted, don't propagate error if delete fails
	syncer.NamespaceClient.TryDeleteInstallNamespace(getInstallNamespace(install, meshInstaller.GetDefaultNamespace()))
	// TODO: this will break if there are more than one installs of a given mesh that depend on the CRB
	// Create a CRB per install?
	if meshInstaller.GetCrbName() != "" {
		return syncer.RbacClient.DeleteCrb(meshInstaller.GetCrbName())
	}
	return nil
}
