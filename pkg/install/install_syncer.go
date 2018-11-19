package install

import (
	"context"

	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/consul"
	"github.com/solo-io/supergloo/pkg/install/helm"
	"k8s.io/client-go/kubernetes"

	kubecore "k8s.io/api/core/v1"
	kuberbac "k8s.io/api/rbac/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	helmlib "k8s.io/helm/pkg/helm"
)

type InstallSyncer struct {
	Kube       *kubernetes.Clientset
	MeshClient v1.MeshClient
}

type MeshInstaller interface {
	GetDefaultNamespace() string
	GetCrbName() string
	GetOverridesYaml(install *v1.Install) string
	DoPostHelmInstall(install *v1.Install, kube *kubernetes.Clientset, releaseName string) error
}

func (syncer *InstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	for _, install := range snap.Installs.List() {
		err := syncer.SyncInstall(ctx, install)
		if err != nil {
			return err
		}
	}
	return nil
}

func (syncer *InstallSyncer) SyncInstall(ctx context.Context, install *v1.Install) error {
	var meshInstaller MeshInstaller
	switch install.MeshType {
	case v1.MeshType_CONSUL:
		meshInstaller = &consul.ConsulInstaller{}
	default:
		return errors.Errorf("Unsupported mesh type %v", install.MeshType)
	}

	if err := syncer.SyncInstallImpl(ctx, install, meshInstaller); err != nil {
		return err
	}
	return syncer.createMesh(install)
}

func (syncer *InstallSyncer) SyncInstallImpl(_ context.Context, install *v1.Install, installer MeshInstaller) error {
	// 1. Setup namespace
	installNamespace, err := syncer.SetupInstallNamespace(install, installer.GetDefaultNamespace())
	if err != nil {
		return err
	}

	// 2. Set up ClusterRoleBinding for that namespace
	// This is not cleaned up when deleting namespace so it may already exist on the system, don't fail
	crbName := installer.GetCrbName()
	if crbName != "" {
		err = syncer.CreateCrbIfNotExist(crbName, installNamespace)
		if err != nil {
			return err
		}
	}

	// 3. Install Consul via helm chart
	releaseName, err := syncer.HelmInstall(install.ChartLocator, installNamespace, installer.GetOverridesYaml(install))
	if err != nil {
		return errors.Wrap(err, "Error installing helm chart")
	}

	// 4. Do any additional steps
	return installer.DoPostHelmInstall(install, syncer.Kube, releaseName)
}

func (syncer *InstallSyncer) SetupInstallNamespace(install *v1.Install, defaultNamespace string) (string, error) {
	installNamespace := getInstallNamespace(install, defaultNamespace)

	err := syncer.createNamespaceIfNotExist(installNamespace) // extract to CRD
	if err != nil {
		return installNamespace, errors.Wrap(err, "Error setting up namespace")
	} else {
		return installNamespace, nil
	}
}

func getInstallNamespace(install *v1.Install, defaultNamespace string) string {
	installNamespace := defaultNamespace
	if install.InstallNamespace != "" {
		installNamespace = install.InstallNamespace
	}
	return installNamespace
}

func (syncer *InstallSyncer) createNamespaceIfNotExist(namespaceName string) error {
	_, err := syncer.Kube.CoreV1().Namespaces().Get(namespaceName, kubemeta.GetOptions{})
	if err == nil {
		// Namespace already exists
		return nil
	}
	_, err = syncer.Kube.CoreV1().Namespaces().Create(getNamespace(namespaceName))
	return err
}

func getNamespace(namespaceName string) *kubecore.Namespace {
	return &kubecore.Namespace{
		ObjectMeta: kubemeta.ObjectMeta{
			Name: namespaceName,
		},
	}
}

func (syncer *InstallSyncer) CreateCrbIfNotExist(crbName string, namespaceName string) error {
	_, err := syncer.Kube.RbacV1().ClusterRoleBindings().Get(namespaceName, kubemeta.GetOptions{})
	if err == nil {
		// crb already exists
		return nil
	}
	_, err = syncer.Kube.RbacV1().ClusterRoleBindings().Create(getCrb(crbName, namespaceName))
	return err
}

func getCrb(crbName string, namespaceName string) *kuberbac.ClusterRoleBinding {
	meta := kubemeta.ObjectMeta{
		Name: crbName,
	}
	subject := kuberbac.Subject{
		Kind:      "ServiceAccount",
		Namespace: namespaceName,
		Name:      "default",
	}
	roleRef := kuberbac.RoleRef{
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
		APIGroup: "rbac.authorization.k8s.io",
	}
	return &kuberbac.ClusterRoleBinding{
		ObjectMeta: meta,
		Subjects:   []kuberbac.Subject{subject},
		RoleRef:    roleRef,
	}
}

func (syncer *InstallSyncer) HelmInstall(chartLocator *v1.HelmChartLocator, installNamespace string, overridesYaml string) (string, error) {
	if chartLocator.GetChartPath() != nil {
		return helmInstallPath(chartLocator.GetChartPath(), installNamespace, overridesYaml)
	} else {
		return "", errors.Errorf("Unsupported kind of chart locator")
	}
}

func helmInstallPath(chartPath *v1.HelmChartPath, installNamespace string, overridesYaml string) (string, error) {
	// helm install
	helmClient, err := helm.GetHelmClient()
	if err != nil {
		return "", err
	}

	installPath, err := helm.LocateChartPathDefault(chartPath.Path)
	if err != nil {
		return "", err
	}
	response, err := helmClient.InstallRelease(
		installPath,
		installNamespace,
		helmlib.ValueOverrides([]byte(overridesYaml)))
	helm.Teardown()
	if err != nil {
		return "", err
	} else {
		return response.Release.Name, nil
	}
}

func (syncer *InstallSyncer) createMesh(install *v1.Install) error {
	mesh := getMeshObject(install)
	_, err := syncer.MeshClient.Write(mesh, clients.WriteOpts{})
	return err
}

func getMeshObject(install *v1.Install) *v1.Mesh {
	return &v1.Mesh{
		Metadata: core.Metadata{
			Name:      install.Metadata.Name,
			Namespace: install.Metadata.Namespace,
		},
		TargetMesh: &v1.TargetMesh{
			MeshType: install.MeshType,
		},
		Encryption: install.Encryption,
	}
}
