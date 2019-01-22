package typed

import (
	"io/ioutil"
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"

	"github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/test/setup"
	"k8s.io/client-go/kubernetes"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type ResourceClientTester interface {
	Description() string
	Skip() bool
	Setup(namespace string) factory.ResourceClientFactory
	Teardown(namespace string)
}

func skipKubeTests() bool {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return true
	}
	return false
}

/*
 *
 * KubeCrd
 *
 */

type KubeRcTester struct {
	Crd crd.Crd
}

func (rct *KubeRcTester) Description() string {
	return "kube-crd"
}

func (rct *KubeRcTester) Skip() bool {
	return skipKubeTests()
}

func (rct *KubeRcTester) Setup(namespace string) factory.ResourceClientFactory {
	if namespace != "" {
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	}
	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeResourceClientFactory{
		Crd:         rct.Crd,
		Cfg:         cfg,
		SharedCache: kube.NewKubeCache(),
	}
}

func (rct *KubeRcTester) Teardown(namespace string) {
	setup.TeardownKube(namespace)
}

/*
 *
 * Consul-KV
 *
 */
type ConsulRcTester struct {
	consulInstance *setup.ConsulInstance
	consulFactory  *setup.ConsulFactory
}

func (rct *ConsulRcTester) Description() string {
	return "consul-kv"
}

func (rct *ConsulRcTester) Skip() bool {
	if os.Getenv("RUN_CONSUL_TESTS") != "1" {
		log.Printf("This test downloads and runs consul and is disabled by default. To enable, set RUN_CONSUL_TESTS=1 in your env.")
		return true
	}
	return false
}

func (rct *ConsulRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.consulFactory, err = setup.NewConsulFactory()
	Expect(err).NotTo(HaveOccurred())
	rct.consulInstance, err = rct.consulFactory.NewConsulInstance()
	Expect(err).NotTo(HaveOccurred())
	err = rct.consulInstance.Run()
	Expect(err).NotTo(HaveOccurred())

	consul, err := api.NewClient(api.DefaultConfig())
	Expect(err).NotTo(HaveOccurred())
	return &factory.ConsulResourceClientFactory{
		Consul:  consul,
		RootKey: namespace,
	}
}

func (rct *ConsulRcTester) Teardown(namespace string) {
	rct.consulInstance.Clean()
	rct.consulFactory.Clean()
}

/*
 *
 * File
 *
 */
type FileRcTester struct {
	rootDir string
}

func (rct *FileRcTester) Description() string {
	return "file-based"
}

func (rct *FileRcTester) Skip() bool {
	return false
}

func (rct *FileRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.rootDir, err = ioutil.TempDir("", "base_test")
	Expect(err).NotTo(HaveOccurred())
	return &factory.FileResourceClientFactory{
		RootDir: rct.rootDir,
	}
}

func (rct *FileRcTester) Teardown(namespace string) {
	os.RemoveAll(rct.rootDir)
}

/*
 *
 * Memory
 *
 */
type MemoryRcTester struct {
	rootDir string
}

func (rct *MemoryRcTester) Description() string {
	return "memory-based"
}

func (rct *MemoryRcTester) Skip() bool {
	return false
}

func (rct *MemoryRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.rootDir, err = ioutil.TempDir("", "base_test")
	Expect(err).NotTo(HaveOccurred())
	return &factory.MemoryResourceClientFactory{
		Cache: memory.NewInMemoryResourceCache(),
	}
}

func (rct *MemoryRcTester) Teardown(namespace string) {}

/*
 *
 * KubeCfgMap
 *
 */
type KubeConfigMapRcTester struct{}

func (rct *KubeConfigMapRcTester) Description() string {
	return "kube-configmap-based"
}

func (rct *KubeConfigMapRcTester) Skip() bool {
	return skipKubeTests()
}

func (rct *KubeConfigMapRcTester) Setup(namespace string) factory.ResourceClientFactory {
	err := setup.SetupKubeForTest(namespace)
	Expect(err).NotTo(HaveOccurred())
	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())
	kube, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeConfigMapClientFactory{
		Clientset: kube,
	}
}

func (rct *KubeConfigMapRcTester) Teardown(namespace string) {
	setup.TeardownKube(namespace)
}

/*
 *
 * KubeSecret
 *
 */
type KubeSecretRcTester struct{}

func (rct *KubeSecretRcTester) Description() string {
	return "kube-secret-based"
}

func (rct *KubeSecretRcTester) Skip() bool {
	return skipKubeTests()
}

func (rct *KubeSecretRcTester) Setup(namespace string) factory.ResourceClientFactory {
	err := setup.SetupKubeForTest(namespace)
	Expect(err).NotTo(HaveOccurred())
	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())
	kube, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeSecretClientFactory{
		Clientset: kube,
	}
}

func (rct *KubeSecretRcTester) Teardown(namespace string) {
	setup.TeardownKube(namespace)
}

/*
 *
 * Vault Secret
 *
 */
type VaultRcTester struct {
	vaultInstance *setup.VaultInstance
	vaultFactory  *setup.VaultFactory
}

func (rct *VaultRcTester) Description() string {
	return "vault-secret-based"
}

func (rct *VaultRcTester) Skip() bool {
	if os.Getenv("RUN_VAULT_TESTS") != "1" {
		log.Printf("This test downloads and runs vault and is disabled by default. To enable, set RUN_VAULT_TESTS=1 in your env.")
		return true
	}
	return false
}

func (rct *VaultRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.vaultFactory, err = setup.NewVaultFactory()
	Expect(err).NotTo(HaveOccurred())
	rct.vaultInstance, err = rct.vaultFactory.NewVaultInstance()
	Expect(err).NotTo(HaveOccurred())
	err = rct.vaultInstance.Run()
	Expect(err).NotTo(HaveOccurred())
	rootKey := "/secret/" + namespace
	cfg := vaultapi.DefaultConfig()
	cfg.Address = "http://127.0.0.1:8200"
	vault, err := vaultapi.NewClient(cfg)
	vault.SetToken(rct.vaultInstance.Token())
	Expect(err).NotTo(HaveOccurred())
	return &factory.VaultSecretClientFactory{
		RootKey: rootKey,
		Vault:   vault,
	}
}

func (rct *VaultRcTester) Teardown(namespace string) {
	rct.vaultInstance.Clean()
	rct.vaultFactory.Clean()
}
