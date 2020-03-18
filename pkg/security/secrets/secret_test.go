package cert_secrets_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ca secrets helper", func() {

	Context("Root CA data", func() {

		It("will fail if the root cert is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Type: cert_secrets.RootCertSecretType,
			}
			_, err := cert_secrets.RootCADataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoRootCertFoundError(secret.ObjectMeta)))
		})

		It("will fail if the private key is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID: {},
				},
				Type: cert_secrets.RootCertSecretType,
			}
			_, err := cert_secrets.RootCADataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoPrivateKeyFoundError(secret.ObjectMeta)))
		})

		It("will return the data if all data is present", func() {
			matchData := &cert_secrets.RootCAData{
				PrivateKey: []byte("private_key"),
				RootCert:   []byte("root_cert"),
			}
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:       matchData.RootCert,
					cert_secrets.RootPrivateKeyID: matchData.PrivateKey,
				},
				Type: cert_secrets.RootCertSecretType,
			}
			data, err := cert_secrets.RootCADataFromSecret(secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(matchData))
		})
	})

	Context("Intermediate CA Data", func() {

		It("will fail if the root cert is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Type: cert_secrets.IntermediateCertSecretType,
			}
			_, err := cert_secrets.IntermediateCADataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoCaKeyFoundError(secret.ObjectMeta)))
		})

		It("will fail if the private key is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.CaPrivateKeyID: {},
				},
				Type: cert_secrets.IntermediateCertSecretType,
			}
			_, err := cert_secrets.IntermediateCADataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoCaCertFoundError(secret.ObjectMeta)))
		})

		It("will fail if the cert chain is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.CaPrivateKeyID: {},
					cert_secrets.CaCertID:       {},
				},
				Type: cert_secrets.RootCertSecretType,
			}
			_, err := cert_secrets.IntermediateCADataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoCertChainFoundError(secret.ObjectMeta)))
		})

		It("will return the data if all data is present", func() {
			matchData := &cert_secrets.IntermediateCAData{
				RootCAData: cert_secrets.RootCAData{
					PrivateKey: []byte("private_key"),
					RootCert:   []byte("root_cert"),
				},
				CertChain:    []byte("cert_chain"),
				CaCert:       []byte("ca_cert"),
				CaPrivateKey: []byte("ca_key"),
			}
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:       matchData.RootCert,
					cert_secrets.RootPrivateKeyID: matchData.PrivateKey,
					cert_secrets.CertChainID:      matchData.CertChain,
					cert_secrets.CaPrivateKeyID:   matchData.CaPrivateKey,
					cert_secrets.CaCertID:         matchData.CaCert,
				},
				Type: cert_secrets.IntermediateCertSecretType,
			}
			data, err := cert_secrets.IntermediateCADataFromSecret(secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(matchData))
		})

		It("can build the correct secret from the root cert data", func() {
			name, namespace := "name", "namespace"
			matchData := &cert_secrets.IntermediateCAData{
				RootCAData: cert_secrets.RootCAData{
					PrivateKey: []byte("private_key"),
					RootCert:   []byte("root_cert"),
				},
				CertChain:    []byte("cert_chain"),
				CaCert:       []byte("ca_cert"),
				CaPrivateKey: []byte("ca_key"),
			}
			matchSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:       matchData.RootCert,
					cert_secrets.RootPrivateKeyID: matchData.PrivateKey,
					cert_secrets.CertChainID:      matchData.CertChain,
					cert_secrets.CaPrivateKeyID:   matchData.CaPrivateKey,
					cert_secrets.CaCertID:         matchData.CaCert,
				},
				Type: cert_secrets.IntermediateCertSecretType,
			}
			Expect(matchData.BuildSecret(name, namespace)).To(Equal(matchSecret))
		})
	})
})
