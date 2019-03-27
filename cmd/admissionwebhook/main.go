package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/supergloo/pkg/webhook"
)

func main() {
	rootCtx := contextutils.WithLogger(context.Background(), "sidecar-injector-webhook")
	webhookServer := webhook.NewSidecarInjectionWebhook(rootCtx)

	// Start server in a separate routine so we can gracefully handle the shutdown
	go func() {
		if err := webhookServer.ListenAndServeTLS("", ""); err != nil {
			// ErrServerClosed signals that the server has been stopped via Close or Shutdown
			if err == http.ErrServerClosed {
				contextutils.LoggerFrom(rootCtx).Info("webhook server has been stopped")
			} else {
				contextutils.LoggerFrom(rootCtx).Fatalf("unexpected error running webhook server: %v", err)
			}
		}
	}()

	// Listen for OS shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	contextutils.LoggerFrom(rootCtx).Info("received OS shutdown signal, shutting down webhook server gracefully...")
	_ = webhookServer.Shutdown(rootCtx)
}
