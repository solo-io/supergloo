package testutils

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/constants"
	mdsetup "github.com/solo-io/supergloo/pkg/meshdiscovery/setup"
	"github.com/solo-io/supergloo/pkg/setup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespaces(kube kubernetes.Interface, namespaceMetadata ...metav1.ObjectMeta) error {
	for _, meta := range namespaceMetadata {
		_, err := kube.CoreV1().Namespaces().Create(&corev1.Namespace{ObjectMeta: meta})
		if err != nil {
			return err
		}
	}
	return nil
}

func RunSuperglooLocally(ctx context.Context, kube kubernetes.Interface, superglooNs, buildVersion, imageRepoPrefix string) error {
	var err *multierror.Error

	// Supergloo requires these environment variables to be set
	err = multierror.Append(err, os.Setenv(constants.PodNamespaceEnvName, superglooNs))
	image := fmt.Sprintf("%s/%s:%s", imageRepoPrefix, constants.SidecarInjectorImageName, buildVersion)
	err = multierror.Append(err, os.Setenv(constants.SidecarInjectorImageNameEnvName, image))
	err = multierror.Append(err, os.Setenv(constants.SidecarInjectorImagePullPolicyEnvName, "Always"))
	if err.ErrorOrNil() != nil {
		return err.ErrorOrNil()
	}

	go func() {
		defer GinkgoRecover()
		err := setup.Main(ctx, func(e error) {
			defer GinkgoRecover()

			// TODO: we should assert errors here, but it returns errors that are expected to happen (e.g. "the object
			//  has been modified; please apply your changes to the latest version and try again"). What to do?
			//Expect(e).NotTo(HaveOccurred())
		})
		Expect(err).NotTo(HaveOccurred())
	}()

	// Start mesh discovery
	go func() {
		defer GinkgoRecover()
		err := mdsetup.Main(ctx, func(e error) {
			defer GinkgoRecover()

			// TODO: we should assert errors here, but see TODO above
			//Expect(e).NotTo(HaveOccurred())
		})
		Expect(err).NotTo(HaveOccurred())
	}()

	// Terminate the remote supergloo and
	DeleteSuperglooPods(kube, superglooNs)

	return nil

}
