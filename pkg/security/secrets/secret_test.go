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

	Context("cert and key", func() {

		It("will fail if the root cert is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Type: cert_secrets.CertAndKeySecretType,
			}
			_, err := cert_secrets.CertAndKeyDataFromSecret(secret)
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
				Type: cert_secrets.CertAndKeySecretType,
			}
			_, err := cert_secrets.CertAndKeyDataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoPrivateKeyFoundError(secret.ObjectMeta)))
		})

		It("will fail if the cert chain is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:   {},
					cert_secrets.PrivateKeyID: {},
				},
				Type: cert_secrets.CertAndKeySecretType,
			}
			_, err := cert_secrets.CertAndKeyDataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoCertChainFoundError(secret.ObjectMeta)))
		})

		It("will return the data if all data is present", func() {
			matchData := &cert_secrets.CertAndKeyData{
				CertChain:  []byte("cert_chain"),
				PrivateKey: []byte("private_key"),
				RootCert:   []byte("root_cert"),
			}
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:   matchData.RootCert,
					cert_secrets.PrivateKeyID: matchData.PrivateKey,
					cert_secrets.CertChainID:  matchData.CertChain,
				},
				Type: cert_secrets.CertAndKeySecretType,
			}
			data, err := cert_secrets.CertAndKeyDataFromSecret(secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(matchData))
		})

		It("can build the correct secret from the cert and key data", func() {
			name, namespace := "name", "namespace"
			matchData := &cert_secrets.CertAndKeyData{
				CertChain:  []byte("cert_chain"),
				PrivateKey: []byte("private_key"),
				RootCert:   []byte("root_cert"),
			}
			matchSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:   matchData.RootCert,
					cert_secrets.PrivateKeyID: matchData.PrivateKey,
					cert_secrets.CertChainID:  matchData.CertChain,
				},
				Type: cert_secrets.CertAndKeySecretType,
			}
			Expect(matchData.BuildSecret(name, namespace)).To(Equal(matchSecret))
		})
	})

	Context("Root Ca Data", func() {

		It("will fail if the root cert is not present", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Type: cert_secrets.RootCertSecretType,
			}
			_, err := cert_secrets.RootCaDataFromSecret(secret)
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
				Type: cert_secrets.RootCertSecretType,
			}
			_, err := cert_secrets.RootCaDataFromSecret(secret)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(cert_secrets.NoCaCertFoundError(secret.ObjectMeta)))
		})

		It("will return the data if all data is present", func() {
			matchData := &cert_secrets.RootCaData{
				CertAndKeyData: cert_secrets.CertAndKeyData{
					CertChain:  []byte("cert_chain"),
					PrivateKey: []byte("private_key"),
					RootCert:   []byte("root_cert"),
				},
				CaCert:       []byte("ca_cert"),
				CaPrivateKey: []byte("ca_key"),
			}
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name",
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:     matchData.RootCert,
					cert_secrets.PrivateKeyID:   matchData.PrivateKey,
					cert_secrets.CertChainID:    matchData.CertChain,
					cert_secrets.CaPrivateKeyID: matchData.CaPrivateKey,
					cert_secrets.CaCertID:       matchData.CaCert,
				},
				Type: cert_secrets.RootCertSecretType,
			}
			data, err := cert_secrets.RootCaDataFromSecret(secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(matchData))
		})

		It("can build the correct secret from the root cert data", func() {
			name, namespace := "name", "namespace"
			matchData := &cert_secrets.RootCaData{
				CertAndKeyData: cert_secrets.CertAndKeyData{
					CertChain:  []byte("cert_chain"),
					PrivateKey: []byte("private_key"),
					RootCert:   []byte("root_cert"),
				},
				CaCert:       []byte("ca_cert"),
				CaPrivateKey: []byte("ca_key"),
			}
			matchSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					cert_secrets.RootCertID:     matchData.RootCert,
					cert_secrets.PrivateKeyID:   matchData.PrivateKey,
					cert_secrets.CertChainID:    matchData.CertChain,
					cert_secrets.CaPrivateKeyID: matchData.CaPrivateKey,
					cert_secrets.CaCertID:       matchData.CaCert,
				},
				Type: cert_secrets.RootCertSecretType,
			}
			Expect(matchData.BuildSecret(name, namespace)).To(Equal(matchSecret))
		})
	})
})
