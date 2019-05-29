package appmesh

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"text/template"

	"github.com/solo-io/go-utils/installutils/helmchart"

	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/cert"
	"k8s.io/helm/pkg/manifest"
)

type AutoInjectionReconciler interface {
	Reconcile(autoInjectionEnabled bool) error
}

type autoInjectionReconciler struct {
	kube                     kubernetes.Interface
	installer                Installer
	superglooNamespace       string
	sidecarInjectorImageName string
}

func NewAutoInjectionReconciler(kube kubernetes.Interface, installer Installer, sgNamespace, sidecarInjectorImage string) AutoInjectionReconciler {
	return &autoInjectionReconciler{
		kube:                     kube,
		installer:                installer,
		superglooNamespace:       sgNamespace,
		sidecarInjectorImageName: sidecarInjectorImage,
	}
}

func (r *autoInjectionReconciler) Reconcile(autoInjectionEnabled bool) error {

	// Retrieve and render manifests for auto-injection resources
	autoInjectionManifests, err := r.renderAutoInjectionManifests()
	if err != nil {
		return errors.Wrapf(err, "failed to generate manifests for auto-injection resources")
	}

	// Remove the auto-injection resources if not needed
	if !autoInjectionEnabled {
		if err := r.installer.Delete(r.superglooNamespace, bytes.NewBufferString(autoInjectionManifests.CombinedString())); err != nil {
			return errors.Wrapf(err, "failed to delete auto-injection resources [%s]", autoInjectionManifests.Names())
		}
	} else {

		secretRelatedManifests, otherManifests := separateSecretManifests(autoInjectionManifests)

		// Ensure secret resources separately. This is necessary in cases
		if err := r.ensureSecrets(r.superglooNamespace, secretRelatedManifests); err != nil {
			return errors.Wrapf(err, "failed reconcile auto-injection secret resources [%s]", secretRelatedManifests.Names())
		}

		// TODO(marco): this currently naively checks if the auto-injection resources all exist: if not, it recreates all of them. Swap in the new installer here when it's ready.
		if err := r.ensureDeployment(r.superglooNamespace, otherManifests); err != nil {
			return errors.Wrapf(err, "failed to reconcile auto-injection resources [%s]", otherManifests.Names())
		}
	}

	return nil
}

func (r *autoInjectionReconciler) renderAutoInjectionManifests() (helmchart.Manifests, error) {

	// Retrieve the config map that contains the manifests for all the auto-injection resources
	cm, err := r.kube.CoreV1().ConfigMaps(r.superglooNamespace).Get(resourcesConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve %s config map", resourcesConfigMapName)
	}

	certs, err := generateSelfSignedCertificate(cert.Config{
		CommonName:   fmt.Sprintf("%s.%s", webhookName, r.superglooNamespace),
		Organization: []string{"solo.io"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				webhookName,
				fmt.Sprintf("%s.%s", webhookName, r.superglooNamespace),
				fmt.Sprintf("%s.%s.svc", webhookName, r.superglooNamespace),
				fmt.Sprintf("%s.%s.svc.cluster.local", webhookName, r.superglooNamespace),
			},
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate certificate for webhook server")
	}

	type Values struct {
		Name,
		Namespace,
		ServerCert,
		ServerCertKey,
		CaBundle string
		Image string
	}

	values := Values{
		Name:          webhookName,
		Namespace:     r.superglooNamespace,
		Image:         r.sidecarInjectorImageName,
		ServerCert:    base64.StdEncoding.EncodeToString(certs.serverCertificate),
		ServerCertKey: base64.StdEncoding.EncodeToString(certs.serverCertKey),
		CaBundle:      base64.StdEncoding.EncodeToString(certs.caCertificate),
	}

	// Each key/value pair in configMap.Data is a manifest
	var autoInjectionManifests []manifest.Manifest
	for name, value := range cm.Data {

		tmpl, err := template.New(name).Parse(value)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing template %s in config map %s", name, resourcesConfigMapName)
		}

		buf := &bytes.Buffer{}
		err = tmpl.Execute(buf, values)
		if err != nil {
			return nil, errors.Wrapf(err, "executing template %s in config map %s", name, resourcesConfigMapName)
		}

		man := manifest.SplitManifests(map[string]string{name: buf.String()})
		autoInjectionManifests = append(autoInjectionManifests, man...)
	}

	return autoInjectionManifests, nil
}

func separateSecretManifests(manifests helmchart.Manifests) (secretRelated, regular helmchart.Manifests) {
	for _, m := range manifests {
		if m.Head.Kind == secretKind || m.Head.Kind == webhookConfigKind {
			secretRelated = append(secretRelated, m)
		} else {
			regular = append(regular, m)
		}
	}
	return
}

// The auto-injection manifest contains two secret related resources:

//   1. A Secret that contains the TLS certificate mounted on the MutatingAdmissionWebhook container
//   2. A MutatingWebhookConfiguration that contains the certificate for the CA that signed the TLS certificate
//
// If both exist, do nothing; if none exists, create both; if any of them does not exist, this means the cluster
// might be in an inconsistent state and we need to clean up and recreate both.
func (r *autoInjectionReconciler) ensureSecrets(namespace string, secretRelatedManifests helmchart.Manifests) error {
	var secretMissing, whConfigMissing bool
	for _, m := range secretRelatedManifests {
		if m.Head.Kind == secretKind {
			secret := &corev1.Secret{}
			if err := yaml.Unmarshal([]byte(m.Content), secret); err != nil {
				return errors.Wrapf(err, "unmarshalling content for manifest %s", m.Name)
			}

			if _, err := r.kube.CoreV1().Secrets(secret.Namespace).Get(secret.Name, metav1.GetOptions{}); err != nil {
				if !k8serrors.IsNotFound(err) {
					return errors.Wrapf(err, "reading %s %s.%s", secretKind, secret.Namespace, secret.Name)
				}
				secretMissing = true
			}
		}
		if m.Head.Kind == webhookConfigKind {

			conf := &admissionv1beta1.MutatingWebhookConfiguration{}
			if err := yaml.Unmarshal([]byte(m.Content), conf); err != nil {
				return errors.Wrapf(err, "unmarshalling content for manifest %s", m.Name)
			}
			_, err := r.kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(conf.Name, metav1.GetOptions{})
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return errors.Wrapf(err, "reading %s =%s", webhookConfigKind, conf.Name)
				}
				whConfigMissing = true
			}
		}
	}

	// If both exist, do nothing
	if !(secretMissing || whConfigMissing) {
		return nil
	}

	// If only one exist, clean up
	if !(secretMissing && whConfigMissing) {
		if err := r.installer.Delete(namespace, bytes.NewBufferString(secretRelatedManifests.CombinedString())); err != nil {
			return errors.Wrapf(err, "failed to delete manifests %s", secretRelatedManifests.Names())
		}
	}

	if err := r.installer.Create(namespace, bytes.NewBufferString(secretRelatedManifests.CombinedString()), 0, false); err != nil {
		return errors.Wrapf(err, "failed to create manifests %s", secretRelatedManifests.Names())
	}
	return nil
}

// TODO: replace with new stateless installer library when ready.
func (r *autoInjectionReconciler) ensureDeployment(namespace string, manifests helmchart.Manifests) error {
	var deploymentMissing, serviceMissing, configMapMissing bool

	for _, m := range manifests {
		if m.Head.Kind == deploymentKind {
			deployment := &appsv1.Deployment{}
			if err := yaml.Unmarshal([]byte(m.Content), deployment); err != nil {
				return errors.Wrapf(err, "unmarshalling content for manifest %s", m.Name)
			}
			if _, err := r.kube.AppsV1().Deployments(deployment.Namespace).Get(deployment.Name, metav1.GetOptions{}); err != nil {
				if !k8serrors.IsNotFound(err) {
					return errors.Wrapf(err, "reading %s %s.%s", deploymentKind, deployment.Namespace, deployment.Name)
				}
				deploymentMissing = true
			}
		}
		if m.Head.Kind == serviceKind {
			service := &corev1.Service{}
			if err := yaml.Unmarshal([]byte(m.Content), service); err != nil {
				return errors.Wrapf(err, "unmarshalling content for manifest %s", m.Name)
			}
			_, err := r.kube.CoreV1().Services(service.Namespace).Get(service.Name, metav1.GetOptions{})
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return errors.Wrapf(err, "reading %s %s.%s", serviceKind, service.Namespace, service.Name)
				}
				serviceMissing = true
			}
		}
		if m.Head.Kind == configMapKind {
			cm := &corev1.ConfigMap{}
			if err := yaml.Unmarshal([]byte(m.Content), cm); err != nil {
				return errors.Wrapf(err, "unmarshalling content for manifest %s", m.Name)
			}
			_, err := r.kube.CoreV1().ConfigMaps(cm.Namespace).Get(cm.Name, metav1.GetOptions{})
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return errors.Wrapf(err, "reading %s %s.%s", configMapKind, cm.Namespace, cm.Name)
				}
				configMapMissing = true
			}
		}
	}

	// If all exist, do nothing
	if !(deploymentMissing || serviceMissing || configMapMissing) {
		return nil
	}

	// If any is missing, clean up
	if !(deploymentMissing && serviceMissing && configMapMissing) {
		if err := r.installer.Delete(namespace, bytes.NewBufferString(manifests.CombinedString())); err != nil {
			return errors.Wrapf(err, "failed to delete manifests %s", manifests.Names())
		}
	}

	if err := r.installer.Create(namespace, bytes.NewBufferString(manifests.CombinedString()), 0, false); err != nil {
		return errors.Wrapf(err, "failed to create manifests %s", manifests.Names())
	}
	return nil
}
