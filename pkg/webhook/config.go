package webhook

import (
	"flag"
	"os"
)

type webhookConfig struct {
	port         int    // port the webhook server listens on
	certFilePath string // path to the x509 server certificate for TLS
	keyFilePath  string // path to the x509 private key matching `CertFile`
}

// Webhook configuration values are read in, order of preference, from:
//   1. flags passed to the go executable
//   2. environment variables
//   3. default location (/etc/webhook/)
func getConfig() webhookConfig {
	defaultCertFile := "/etc/webhook/cert.pem"
	defaultKeyFile := "/etc/webhook/key.pem"

	if envCertFile := os.Getenv("TLS_CERT_PATH"); envCertFile != "" {
		defaultCertFile = envCertFile
	}
	if envKeyFile := os.Getenv("TLS_KEY_PATH"); envKeyFile != "" {
		defaultKeyFile = envKeyFile
	}

	var config webhookConfig
	flag.IntVar(&config.port, "port", 443, "Webhook server port.")
	flag.StringVar(&config.certFilePath, "tls-cert-file", defaultCertFile, "File containing the PEM-encoded x509 server certificate for HTTPS.")
	flag.StringVar(&config.keyFilePath, "tls-key-file", defaultKeyFile, "File containing the x509 private key to --tls-cert-file.")
	flag.Parse()

	return config
}
