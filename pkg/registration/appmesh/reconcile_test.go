package appmesh

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/version"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Reconciler", func() {

	var (
		kube kubernetes.Interface
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)

		kube = fake.NewSimpleClientset()

		// Create the config map containing the manifests for the auto-injection resources
		configMap := &corev1.ConfigMap{}
		cm, err := ioutil.ReadFile("test/test-sidecar-injector-configmap.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(yaml.Unmarshal(cm, configMap)).NotTo(HaveOccurred())
		_, err = kube.CoreV1().ConfigMaps(testNamespace).Create(configMap)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
		Expect(kube.CoreV1().ConfigMaps(testNamespace).Delete(resourcesConfigMapName, &metav1.DeleteOptions{})).NotTo(HaveOccurred())
	})

	Describe("render auto-injection resources config map", func() {

		It("correctly renders the manifest templates contained in the config map", func() {
			reconciler := autoInjectionReconciler{kube: kube}
			autoInjectionManifests, err := reconciler.renderAutoInjectionManifests(testNamespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(autoInjectionManifests).To(HaveLen(5))

			for _, m := range autoInjectionManifests {
				switch m.Name {
				case "secret.tpl":
					secret := &corev1.Secret{}
					Expect(yaml.Unmarshal([]byte(m.Content), secret)).NotTo(HaveOccurred())
					Expect(secret.Namespace).To(BeEquivalentTo(testNamespace))
					Expect(secret.Name).To(BeEquivalentTo(webhookName))
					Expect(secret.Data).To(HaveLen(2))

				case "deployment.tpl":
					deployment := &appsv1.Deployment{}
					Expect(yaml.Unmarshal([]byte(m.Content), deployment)).NotTo(HaveOccurred())
					Expect(deployment.Namespace).To(BeEquivalentTo(testNamespace))
					Expect(deployment.Name).To(BeEquivalentTo(webhookName))
					containers := deployment.Spec.Template.Spec.Containers
					Expect(containers).To(HaveLen(1))
					Expect(containers[0].Image).To(BeEquivalentTo(fmt.Sprintf("%s:%s", webhookImageName, version.GetWebhookImageTag())))
					Expect(containers[0].ImagePullPolicy).To(BeEquivalentTo(webhookImagePullPolicy))
					volumes := deployment.Spec.Template.Spec.Volumes
					Expect(volumes).To(HaveLen(1))
					Expect(volumes[0].Secret.SecretName).To(BeEquivalentTo(webhookName))

				case "service.tpl":
					service := &corev1.Service{}
					Expect(yaml.Unmarshal([]byte(m.Content), service)).NotTo(HaveOccurred())
					Expect(service.Namespace).To(BeEquivalentTo(testNamespace))
					Expect(service.Name).To(BeEquivalentTo(webhookName))
					Expect(service.Spec.Ports).To(ContainElement(corev1.ServicePort{Port: 443, TargetPort: intstr.FromInt(443)}))

				case "config-map.tpl":
					configMap := &corev1.ConfigMap{}
					Expect(yaml.Unmarshal([]byte(m.Content), configMap)).NotTo(HaveOccurred())
					Expect(configMap.Namespace).To(BeEquivalentTo(testNamespace))
					Expect(configMap.Name).To(BeEquivalentTo(webhookName))
					Expect(configMap.Data).To(HaveLen(1))

				case "mutating-webhook-configuration.tpl":
					config := &v1beta1.MutatingWebhookConfiguration{}
					Expect(yaml.Unmarshal([]byte(m.Content), config)).NotTo(HaveOccurred())
					Expect(config.Name).To(BeEquivalentTo(webhookName))
					Expect(config.Webhooks).To(HaveLen(1))
					webhook := config.Webhooks[0]
					Expect(webhook.Name).To(BeEquivalentTo(fmt.Sprintf("%s.%s.svc", webhookName, testNamespace)))
					Expect(webhook.ClientConfig.Service.Namespace).To(BeEquivalentTo(testNamespace))
					Expect(webhook.ClientConfig.Service.Name).To(BeEquivalentTo(webhookName))
					Expect(len(webhook.ClientConfig.CABundle)).To(BeNumerically(">", 1))
				}
			}
		})

		It("fails if it cannot find the config map", func() {
			reconciler := autoInjectionReconciler{kube: kube}
			_, err := reconciler.renderAutoInjectionManifests("some-ns")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("test reconciliation", func() {

		Context("auto-injection is disabled", func() {
			It("deletes all auto-injection resources", func() {
				mockInstaller := NewMockInstaller(ctrl)
				mockInstaller.EXPECT().Delete(testNamespace, gomock.Any()).Return(nil).Times(1)

				err := NewAutoInjectionReconciler(kube, mockInstaller).Reconcile(false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("auto-injection is enabled", func() {

			Context("none of the auto-injection components exist", func() {
				It("creates all components", func() {
					mockInstaller := NewMockInstaller(ctrl)
					mockInstaller.EXPECT().Delete(testNamespace, gomock.Any()).Return(nil).Times(0)
					// One call to Create for secret resources and one for the non-secret ones
					mockInstaller.EXPECT().Create(testNamespace, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

					err := NewAutoInjectionReconciler(kube, mockInstaller).Reconcile(true)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("only secret-related resources exist", func() {
				It("non-secret resources will be created", func() {
					_, err := kube.CoreV1().Secrets(testNamespace).Create(secret)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(whConfig)
					Expect(err).NotTo(HaveOccurred())

					mockInstaller := NewMockInstaller(ctrl)
					mockInstaller.EXPECT().Delete(testNamespace, gomock.Any()).Return(nil).Times(0)
					// One call to Create for non-secret resources
					mockInstaller.EXPECT().Create(testNamespace, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

					err = NewAutoInjectionReconciler(kube, mockInstaller).Reconcile(true)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("secret-related resources are in an inconsistent state", func() {
				It("secret resources will be recreated", func() {
					_, err := kube.CoreV1().Secrets(testNamespace).Create(secret)
					Expect(err).NotTo(HaveOccurred())

					mockInstaller := NewMockInstaller(ctrl)
					// Clean up secret resources
					mockInstaller.EXPECT().Delete(testNamespace, gomock.Any()).Return(nil).Times(1)
					// Create both
					mockInstaller.EXPECT().Create(testNamespace, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

					err = NewAutoInjectionReconciler(kube, mockInstaller).Reconcile(true)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("all resources already exist", func() {
				It("secret resources will be recreated", func() {
					_, err := kube.CoreV1().Secrets(testNamespace).Create(secret)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(whConfig)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.AppsV1().Deployments(testNamespace).Create(deployment)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.CoreV1().Services(testNamespace).Create(service)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.CoreV1().ConfigMaps(testNamespace).Create(configMap)
					Expect(err).NotTo(HaveOccurred())

					mockInstaller := NewMockInstaller(ctrl)
					mockInstaller.EXPECT().Delete(testNamespace, gomock.Any()).Return(nil).Times(0)
					mockInstaller.EXPECT().Create(testNamespace, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)

					err = NewAutoInjectionReconciler(kube, mockInstaller).Reconcile(true)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("some of the non-secret resources do not exist", func() {
				It("secret resources will be recreated", func() {
					_, err := kube.CoreV1().Secrets(testNamespace).Create(secret)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(whConfig)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.AppsV1().Deployments(testNamespace).Create(deployment)
					Expect(err).NotTo(HaveOccurred())

					mockInstaller := NewMockInstaller(ctrl)
					// Clean up non-secret resources
					mockInstaller.EXPECT().Delete(testNamespace, gomock.Any()).Return(nil).Times(1)
					// Recreate non-secret resources
					mockInstaller.EXPECT().Create(testNamespace, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

					err = NewAutoInjectionReconciler(kube, mockInstaller).Reconcile(true)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("all resources are in an inconsistent state", func() {
				It("secret resources will be recreated", func() {
					_, err := kube.CoreV1().Secrets(testNamespace).Create(secret)
					Expect(err).NotTo(HaveOccurred())
					_, err = kube.AppsV1().Deployments(testNamespace).Create(deployment)
					Expect(err).NotTo(HaveOccurred())

					mockInstaller := NewMockInstaller(ctrl)
					// Clean up both
					mockInstaller.EXPECT().Delete(testNamespace, gomock.Any()).Return(nil).Times(2)
					// Recreate both
					mockInstaller.EXPECT().Create(testNamespace, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

					err = NewAutoInjectionReconciler(kube, mockInstaller).Reconcile(true)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})

var secret = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      webhookName,
		Namespace: testNamespace,
	},
}

var whConfig = &v1beta1.MutatingWebhookConfiguration{
	ObjectMeta: metav1.ObjectMeta{
		Name: webhookName,
	},
}

var deployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      webhookName,
		Namespace: testNamespace,
	},
}

var service = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      webhookName,
		Namespace: testNamespace,
	},
}

var configMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      webhookName,
		Namespace: testNamespace,
	},
}
