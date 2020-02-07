package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("secret writer", func() {

	var (
		fakeClientset *fake.Clientset
		secretWriter  common.SecretWriter

		testErr   = eris.New("hello")
		namespace = "hello"
	)

	BeforeEach(func() {
		fakeClientset = fake.NewSimpleClientset()
		secretWriter = common.DefaultSecretWriterProvider(fakeClientset, namespace)
	})

	It("will attempt to update if create fails with already exist", func() {
		testSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "hello",
			},
		}
		fakeClientset.PrependReactor("update", "secrets",
			func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, testErr
			})
		fakeClientset.CoreV1().Secrets(namespace).Create(testSecret)
		err := secretWriter.Apply(testSecret)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(testErr))

	})

	It("will return result if create fails with any other error", func() {
		testSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "hello",
			},
		}
		err := secretWriter.Apply(testSecret)
		Expect(err).NotTo(HaveOccurred())
		_, err = fakeClientset.CoreV1().Secrets(namespace).Get(namespace, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
})
