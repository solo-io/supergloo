package translation_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/translation"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CertIssueTranslator", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		mockSecretClient *corev1clients.MockSecretClient
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockSecretClient = corev1clients.NewMockSecretClient(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	XIt("will do nothing if no signing cert is present", func() {

		translator := translation.NewTranslator(mockSecretClient)

		rootCaData := &secrets.RootCAData{
			PrivateKey: []byte(rootKey),
			RootCert:   []byte(rootCert),
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "i'm a",
				Namespace: "secret",
			},
			Data: rootCaData.ToSecretData(),
		}

		certRequest := &certificatesv1.CertificateRequest{
			Spec: certificatesv1.CertificateRequestSpec{
				CertificateSigningRequest: []byte(csr),
			},
		}

		issuedCert := &certificatesv1.IssuedCertificate{
			Spec: certificatesv1.IssuedCertificateSpec{
				CertificateAuthority: &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
					GlooMeshCa: &certificatesv1.GlooMeshCA{
						Signer: &certificatesv1.GlooMeshCA_SigningCertificateSecret{
							SigningCertificateSecret: ezkube.MakeObjectRef(secret),
						},
					},
				},
			},
		}

		mockSecretClient.EXPECT().GetSecret(ctx, ezkube.MakeClientObjectKey(secret)).Return(secret, nil)

		output, err := translator.Translate(ctx, certRequest, issuedCert)
		Expect(err).NotTo(HaveOccurred())

		translatedCertByt, _ := pem.Decode(output.SignedCertificate)
		translatedCert, err := x509.ParseCertificate(translatedCertByt.Bytes)
		Expect(err).NotTo(HaveOccurred())

		precomputedCertByt, _ := pem.Decode([]byte(outputCert))
		precomputedCert, err := x509.ParseCertificate(precomputedCertByt.Bytes)
		Expect(err).NotTo(HaveOccurred())

		Expect(precomputedCert.Extensions).To(Equal(translatedCert.Extensions))
		Expect(precomputedCert.Issuer).To(Equal(translatedCert.Issuer))
		Expect(precomputedCert.Subject).To(Equal(translatedCert.Subject))
		Expect(err).NotTo(HaveOccurred())
	})
})

const (
	rootKey = `-----BEGIN PRIVATE KEY-----
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
uqlk3Vr0zwtxhEdvBV6joOJAh143AnY=
-----END PRIVATE KEY-----
`
	rootCert = `-----BEGIN CERTIFICATE-----
MIIFRjCCAy6gAwIBAgIUNsPXGLirqhAjEc80yxfnwnxhxfgwDQYJKoZIhvcNAQEL
BQAwIzEhMB8GA1UEAwwYZW50ZXJwcmlzZS1uZXR3b3JraW5nLWNhMB4XDTIxMDUw
NDE5MTU1N1oXDTMxMDUwMjE5MTU1N1owIzEhMB8GA1UEAwwYZW50ZXJwcmlzZS1u
ZXR3b3JraW5nLWNhMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEApSz+
lHqzpDF2HCLGHNM7P3ZaEDlD0KKvoB6TjiKUtuO6mfQTmo1wMYmLRqn/y8lm9dN0
zk5jI3HFjsPBaTY5PoBO4rAW1hlN6sNHtY9ffm6kEpwbPF/Wf2z7HfFgl8/r36iy
stDKHsdGLn0sPftng4b2BbZvaT+w78ROoJUebEL1abaFKuJxoSWhHEAJqXuePji4
zp1Q0V031KyZ/QX3GF9B/u6qkyJBnpfxLFdWLVYyl4OnQyev2E4GH5px0x1fAFak
/Y7cMoG93+NXP9uVFx7I/cqJLplq5SJA57wFvmo920+LJcKT9ka1LODzhlzUKKL+
xUTV558Di+I/0nhe5LAt4At4iPPjYVmdRr65hJ9xIUZO0i83X2UHN3NOV/jeT/tj
dcCm9WlLguUAcBxDEST2Ve6laHncm02xaY2so+/MQzCRLRyfsp+J74Le9g0hx1TD
FhsJ/Xs8UW+TkiK5aO+jwgsNZC3X0igGkwjnJWbfvMllZ6QkBZ3CCRD7IoKubmAL
SVFydN8sngSS7eZ1zlUxAwYywgT/9IwUpcccnTDJxpMXlL7fJAUixOmm4niqI8dN
qLFUZZmTiyGUBLlDPZjCnTNCZxHlsPFzd2kz608xQzwqvTvaL4mwmhZHD6bKY5bY
qP3Ln7AC92xaqPr5FLh4xVFY0SnRzwBfQSGj2b8CAwEAAaNyMHAwHQYDVR0OBBYE
FM9O99WdQ0A7Xiktx61pjREZAiA4MB8GA1UdIwQYMBaAFM9O99WdQ0A7Xiktx61p
jREZAiA4MA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsG
AQUFBwMBMA0GCSqGSIb3DQEBCwUAA4ICAQAMPzg0fF1F67UF03/UcBzkVAYMahgL
ExO/J9odXp8zCF5hS5pUiKH/LmJ3IlI1RxjCxEWNyYT7IMIZeD4laqCy3MKJXb0p
QuxSLTFHUHC6c7r7jbr0ApuDoUOxNkbI2vkqY+2dwgHFSOP4Yu3q9Fz2Dy0so+Kb
Ss++N1ZUqQB8AEfl3gKpbZIqaIlqyifr1UghmfrJbveK92KDzE5Ac1NiOxWY/bHH
lousr1fLGx/nk1JCuGkMB+rXjW1HUjUFc7TqJgjlZu2voOM5V2r3izgeizZWwcbB
QgbV/hYJRCEj4DnLz9mRnrlntdbR/0lqsFKVYO/R9qXIZlVvfpN+UPFGK/Id6IAI
6BACy/0JvTXsQcvMpU3u+tp3UkqK5Y9+WOoM26GioveLN+EadCWZYwb1aYAG+xIn
eoxqcXKtZs3+uq5NoL3FhVEepEydW/AH54U/URIk2CSJqNF27lrFNih40iOzn9kP
M+8jl2Gk3jPEVSD7bIB6PK/ysX7McLmEMPaafPZ2RebkXAprz08e+G+tULQzC8BV
6gsrMoPjSAg5FE7TmAcWsdjIJ/hkkptxNd/mbNg1ikHwp068+YA3/TTXfSipPjSA
x8fLXGqGq3eYyF47JIchqEiRuFovC0UATmJxu8Jts2L0WQlruz4VcNaf6bUAZw0c
R9g4Q5ITvVNLvg==
-----END CERTIFICATE-----
`
	csr = `-----BEGIN CERTIFICATE REQUEST-----
MIICyDCCAbACAQAwIzEhMB8GA1UEAwwYZW50ZXJwcmlzZS1uZXR3b3JraW5nLWNh
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA+Ros/oQM2hkjHZ8MnThi
UH4czcHoEQk5qas4ob2/ALIRMgUICjxUFYrjG+cFgOAYNlR0ST/BMz+eZ0oW6u+0
o1jhYW+MWROfoNlqLK6bHb3iRpXTZ4a3t/JXVnDE45KyfbuFvbyGuR+gCb4hTSgv
evJ3djagUMcFbxYoCy6lQlsOVaWd/S45Jn+bDqJucZDLVRE0jhVjXm/+YawoZqAm
sEkPdVW+JNyy/FxbKofUFI3y9nYvEIq9kE2d9G2AjlpwAZP9j/CSeDVw05JuTubf
fsiNdVdLb2WTyI1OiFSPBdJAJrvuXGJI8uznM9UaO9cOqm2eyBsX/02tOQn3Pcky
0QIDAQABoGAwXgYJKoZIhvcNAQkOMVEwTzAJBgNVHRMEAjAAMAsGA1UdDwQEAwIF
oDAdBgNVHSUEFjAUBggrBgEFBQcDAgYIKwYBBQUHAwEwFgYDVR0RBA8wDYILKi5n
bG9vLW1lc2gwDQYJKoZIhvcNAQELBQADggEBAFFqD4zdip9ZLezUM4zi1ofWefGd
eT5sXH++P6minW7zpfDQIIce26mvzTTpD2VR3xXAVoDBROM8Nu8tBcBKEsgoEAqL
N9FMVGEJ8VFmpEM6UuPm2XGBLxDL0s2sXK7gTf62RhrjUEIa0Sh+crRU8rSUOA0l
tUg2nnhPwABsXNPxGvc7hn31ufJOd12l5Q14hypZ13HWxYeberOCXtOzuLqGLjrG
sI1qfty53Ppa0SETviT6J/K0fXxHctMQbRQPNZP2+l1XzvfkUHSyw3FNDQQPwf8w
vT6dljtYarF+6oELQbYWRm69f3WW3LMk/KXRlcXl3v+T5zcgl2QISYGHcyo=
-----END CERTIFICATE REQUEST-----
`
	outputCert = `-----BEGIN CERTIFICATE-----
MIIEIDCCAgigAwIBAgIQAVe23lms5XZLrNetW6O7HTANBgkqhkiG9w0BAQsFADAj
MSEwHwYDVQQDDBhlbnRlcnByaXNlLW5ldHdvcmtpbmctY2EwHhcNMjEwNTA1MjAw
OTI4WhcNMjIwNTA1MjAwOTI4WjAAMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEA+Ros/oQM2hkjHZ8MnThiUH4czcHoEQk5qas4ob2/ALIRMgUICjxUFYrj
G+cFgOAYNlR0ST/BMz+eZ0oW6u+0o1jhYW+MWROfoNlqLK6bHb3iRpXTZ4a3t/JX
VnDE45KyfbuFvbyGuR+gCb4hTSgvevJ3djagUMcFbxYoCy6lQlsOVaWd/S45Jn+b
DqJucZDLVRE0jhVjXm/+YawoZqAmsEkPdVW+JNyy/FxbKofUFI3y9nYvEIq9kE2d
9G2AjlpwAZP9j/CSeDVw05JuTubffsiNdVdLb2WTyI1OiFSPBdJAJrvuXGJI8uzn
M9UaO9cOqm2eyBsX/02tOQn3Pcky0QIDAQABo3MwcTAOBgNVHQ8BAf8EBAMCAgQw
DwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU7U7ME8d9UpIRb9j3Cse0r2CwQOEw
HwYDVR0jBBgwFoAUz0731Z1DQDteKS3HrWmNERkCIDgwDgYDVR0RAQH/BAQwAoIA
MA0GCSqGSIb3DQEBCwUAA4ICAQAbIFFoHtGNA4EqSBQNWTjWgFnjgy8X68Bl6s1d
K3G54hrlzup9cFTUyk1y4qSXVFUpN1nzHhriixQFyu3k/zR5y0e4fukt38oot/Rf
KRfQ5ZssTlUTj0lDnXIyI3hGwyAwVDDADQ8NZpc8uck2krCoOf8Kgq7UKNf6zL5g
m8j881uQkdyXNcS19/s9YeWv+YgFGca55G0xfxgzsnJ29NERmOfoYxtg2IYZaN73
gbkCwFrncFyOD8y7OmZvmtmZ+mhyZmADusCWUS52FEf/upDPQsLYojM58gTXEis/
8g3/qMvGdzODIrb8Gc0lwOaC5lGVV3st1s9kx50zK5Z+w8UAte+dkkAfe1CqrxP7
J09qEphnMS/lF9I0xZ+rMfacc4UVTzd8f88IcS4TMXueFvb6Ja3SAo99KOi/hUwh
hio5FIf4PYXh5+mVJuWe6CsbIAsmnev0hyIA+LTKaW7GaCLKpB7zGOrWaursj3fe
3JF6aW0ZyJPJ6r+7ZtVBC0wI0Pzxg/0lPS46DJasUooWPFTAajvuNYBEzYB0peVw
VhtE8pfas2qpIxrSNVl9w+3RzR9+PfKGtUSxVvQC8YKBvsJXqtxNgrCdhE0s304Y
fUD178RKaIpfm8YE5x1BChvY1tLkxJlnk5fw0+/XAHfqT9yT4aRugXlo0Ic0lB+t
XmsVWw==
-----END CERTIFICATE-----
`
)
