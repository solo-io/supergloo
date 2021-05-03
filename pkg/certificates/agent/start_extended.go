package agent

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/skv2/pkg/bootstrap"
)

// Options for extending the functionality of the Networking controller
type ExtensionOpts struct {
	NetworkingReconciler CertAgentReconcilerExtensionOpts
}

type MakeExtensionOpts func(ctx context.Context, parameters bootstrap.StartParameters) ExtensionOpts

func (opts *ExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {
	opts.NetworkingReconciler.initDefaults(parameters)
}

// Options for overriding functionality of the Networking Reconciler
type CertAgentReconcilerExtensionOpts struct {

	// Hook to override Translator used by Networking Reconciler
	MakeTranslator func(translator translation.Translator) translation.Translator
}

func (opts *CertAgentReconcilerExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {

	if opts.MakeTranslator == nil {
		// use default translator
		opts.MakeTranslator = func(translator translation.Translator) translation.Translator {
			return translator
		}
	}
}
