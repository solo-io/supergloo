package testutils

import (
	"log"
	"os"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	v1alpha12 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

func MustKubeClient() kubernetes.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return kubeClient
}

func MustApiExtsClient() apiexts.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	apiExtsClient, err := apiexts.NewForConfig(restConfig)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return apiExtsClient
}

// call with a var _ = RegisterCrdsLockingSuite
// will fail tests if there's alraedy a BeforeSuite/AfterSuite
func RegisterCrdsLockingSuite() struct{} {

	var (
		kube kubernetes.Interface
		lock *clusterlock.TestClusterLocker
		exts apiexts.Interface
	)

	var _ = BeforeSuite(func() {
		kube = MustKubeClient()
		var err error

		lock, err = clusterlock.NewTestClusterLocker(kube, clusterlock.Options{
			IdPrefix: os.ExpandEnv("superglooe2e-{$BUILD_ID}-"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock(retry.OnRetry(func(n uint, err error) {
			log.Printf("waiting to acquire lock with err: %v", err)
		}))).NotTo(HaveOccurred())

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		exts, err = apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		err = v1alpha1.RbacConfigCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha1.ServiceRoleCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha1.ServiceRoleBindingCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha12.MeshPolicyCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha3.DestinationRuleCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha3.VirtualServiceCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
	})

	var _ = AfterSuite(func() {
		defer lock.ReleaseLock()

		err := exts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(v1alpha1.RbacConfigCrd.FullName(), nil)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha1.ServiceRoleCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha1.ServiceRoleBindingCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha12.MeshPolicyCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha3.DestinationRuleCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
		err = v1alpha3.VirtualServiceCrd.Register(exts)
		Expect(err).NotTo(HaveOccurred())
	})

	return struct{}{}
}
