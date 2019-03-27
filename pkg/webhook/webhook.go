package webhook

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/solo-io/supergloo/pkg/webhook/clients"

	"github.com/solo-io/go-utils/contextutils"
)

const HandlePattern = "/pods/inject-sidecar"

type AdmissionWebhookServer struct {
	*http.Server
}

func NewSidecarInjectionWebhook(ctx context.Context) *AdmissionWebhookServer {
	logger := contextutils.LoggerFrom(ctx)

	// Initialize global client set
	if err := clients.InitClientSet(ctx); err != nil {
		logger.Fatalf("failed to create webhook client set: %v", err)
	}

	// Register component that handles sidecar injection
	RegisterSidecarInjectionHandler()

	// Retrieve the webhook configuration
	config := getConfig()
	logger.Debugf("server config is: %v", config)

	// Load key pair for TLS config
	keyPair, err := tls.LoadX509KeyPair(config.certFilePath, config.keyFilePath)
	if err != nil {
		logger.Fatalf("failed to load key pair: %v", err)
	}
	logger.Debug("successfully loaded key pair")

	// Register request handler
	mux := http.NewServeMux()
	mux.HandleFunc(HandlePattern, handlePodCreation)

	return &AdmissionWebhookServer{
		&http.Server{
			Addr:      fmt.Sprintf(":%v", config.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{keyPair}},
			Handler:   mux,
		},
	}
}
