package mesh_networking

import (
	"context"
	"time"

	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/pkg/ezkube"

	certissuerinput "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	certissuerreconciliation "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	skinput "github.com/solo-io/skv2/contrib/pkg/input"
)

// Options for extending the functionality of the Networking controller
type ExtensionOpts struct {
	NetworkingReconciler NetworkingReconcilerExtensionOpts

	CertIssuerReconciler CertIssuerReconcilerExtensionOpts
}

type MakeExtensionOpts func(parameters bootstrap.StartParameters) ExtensionOpts

func (opts *ExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {
	opts.NetworkingReconciler.initDefaults(parameters)
	opts.CertIssuerReconciler.initDefaults(parameters)
}

// Options for overriding functionality of the Networking Reconciler
type NetworkingReconcilerExtensionOpts struct {

	// Hook to override how the Cert Issuer Reconciler is registered (defaults to the multi cluster manager)
	RegisterNetworkingReconciler reconciliation.RegisterReconcilerFunc

	// Hook to override the User Snapshot Builder used by Networking Reconciler
	MakeUserSnapshotBuilder func(params bootstrap.StartParameters) input.RemoteBuilder

	// Hook to override Translator used by Networking Reconciler
	MakeTranslator func(translator translation.Translator) translation.Translator

	// Hook to override how the Networking Reconciler applies output snapshots
	SyncNetworkingOutputs reconciliation.SyncOutputsFunc
}

func (opts *NetworkingReconcilerExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {
	if opts.RegisterNetworkingReconciler == nil {
		// use default translator
		opts.RegisterNetworkingReconciler = func(ctx context.Context, reconcile skinput.SingleClusterReconcileFunc, reconcileOpts input.ReconcileOptions) (skinput.InputReconciler, error) {
			return input.RegisterInputReconciler(
				ctx,
				parameters.Clusters,
				func(id ezkube.ClusterResourceId) (bool, error) {
					return reconcile(id)
				},
				parameters.MasterManager,
				reconcile,
				reconcileOpts,
			)
		}
	}
	if opts.MakeTranslator == nil {
		// use default translator
		opts.MakeTranslator = func(translator translation.Translator) translation.Translator {
			return translator
		}
	}
	if opts.SyncNetworkingOutputs == nil {
		// sync outputs to multicluster clients (default)
		opts.SyncNetworkingOutputs = func(
			ctx context.Context,
			outputSnap *translation.Outputs,
			errHandler output.ErrorHandler,
		) error {
			return outputSnap.ApplyMultiCluster(ctx, parameters.MasterManager.GetClient(), parameters.McClient, errHandler)
		}
	}
	if opts.MakeUserSnapshotBuilder == nil {
		// read from multicluster clients (default)
		opts.MakeUserSnapshotBuilder = func(_ bootstrap.StartParameters) input.RemoteBuilder {
			return input.NewMultiClusterRemoteBuilder(
				parameters.Clusters,
				parameters.McClient,
			)
		}
	}
}

// Options for overriding functionality of the Cert Issuer
type CertIssuerReconcilerExtensionOpts struct {

	// Hook to override how the Cert Issuer Reconciler is registered (defaults to the multi cluster manager)
	RegisterCertIssuerReconciler certissuerreconciliation.RegisterReconcilerFunc

	// Hook to override the Cert Issuer Snapshot Builder used by Cert Issuer Reconciler
	MakeCertIssuerSnapshotBuilder func(params bootstrap.StartParameters) certissuerinput.Builder

	// Hook to override how the Cert Issuer Reconciler syncs the status of inputs (CertificateRequests)
	SyncCertificateIssuerInputStatuses certissuerreconciliation.SyncStatusFunc
}

func (opts *CertIssuerReconcilerExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {
	if opts.MakeCertIssuerSnapshotBuilder == nil {
		// read from multicluster clients (default)
		opts.MakeCertIssuerSnapshotBuilder = func(_ bootstrap.StartParameters) certissuerinput.Builder {
			return certissuerinput.NewMultiClusterBuilder(
				parameters.Clusters,
				parameters.McClient,
			)
		}
	}
	if opts.RegisterCertIssuerReconciler == nil {
		// initialize cert issuer with multicluster clients (default)
		opts.RegisterCertIssuerReconciler = func(
			ctx context.Context,
			reconcile skinput.MultiClusterReconcileFunc,
			reconcileInterval time.Duration,
		) {
			certissuerinput.RegisterMultiClusterReconciler(
				ctx,
				parameters.Clusters,
				reconcile,
				reconcileInterval,
				certissuerinput.ReconcileOptions{},
			)

		}
	}
	if opts.SyncCertificateIssuerInputStatuses == nil {
		// sync statuses to multicluster clients (default)
		opts.SyncCertificateIssuerInputStatuses = func(ctx context.Context, snapshot certissuerinput.Snapshot) error {
			return snapshot.SyncStatusesMultiCluster(ctx, parameters.McClient, certissuerinput.SyncStatusOptions{
				CertificatesMeshGlooSoloIov1Alpha2CertificateRequest: true,
			})
		}
	}
}
