package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/skv2/contrib/pkg/output/errhandlers"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	corev1client "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/agent/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/agent/output/certagent"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/agent/utils"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/common/secrets"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// all output resources are labeled to prevent
	// resource collisions & garbage collection of
	// secrets the agent doesn't own
	agentLabels = map[string]string{
		fmt.Sprintf("agent.%v", v1alpha2.SchemeGroupVersion.Group): defaults.GetPodNamespace(),
	}

	privateKeySecretType        = corev1.SecretType(fmt.Sprintf("%s/generated_private_key", v1alpha2.SchemeGroupVersion.Group))
	issuedCertificateSecretType = corev1.SecretType(fmt.Sprintf("%s/issued_certificate", v1alpha2.SchemeGroupVersion.Group))
)

type certAgentReconciler struct {
	ctx         context.Context
	builder     input.Builder
	localClient client.Client
}

func Start(
	ctx context.Context,
	builder input.Builder,
	mgr manager.Manager,
) error {
	d := &certAgentReconciler{
		ctx:         ctx,
		builder:     builder,
		localClient: mgr.GetClient(),
	}

	_, err := input.RegisterSingleClusterReconciler(ctx, mgr, d.reconcile, time.Second/2, reconcile.Options{})
	return err
}

const (
	// the map key used to store the agent's private key in a kube Secret
	privateKeySecretKey = "private-key"
)

// reconcile global state
func (r *certAgentReconciler) reconcile(_ ezkube.ResourceId) (bool, error) {
	inputSnap, err := r.builder.BuildSnapshot(r.ctx, "cert-agent", input.BuildOptions{})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	outputs := certagent.NewBuilder(r.ctx, "agent")

	// process issued certificates
	for _, issuedCertificate := range inputSnap.IssuedCertificates().List() {
		if err := r.reconcileIssuedCertificate(
			issuedCertificate,
			inputSnap.Secrets(),
			inputSnap.Pods(),
			inputSnap.CertificateRequests(),
			outputs,
		); err != nil {
			issuedCertificate.Status.Error = err.Error()
			issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_FAILED
		}
	}
	outSnap, err := outputs.BuildSinglePartitionedSnapshot(agentLabels)
	if err != nil {
		return false, err
	}

	errHandler := errhandlers.AppendingErrHandler{}
	outSnap.ApplyLocalCluster(r.ctx, r.localClient, errHandler)

	errs := errHandler.Errors()
	if err := inputSnap.SyncStatuses(r.ctx, r.localClient); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

func (r *certAgentReconciler) reconcileIssuedCertificate(
	issuedCertificate *v1alpha2.IssuedCertificate,
	inputSecrets corev1sets.SecretSet,
	inputPods corev1sets.PodSet,
	inputCertificateRequests v1alpha2sets.CertificateRequestSet,
	outputs certagent.Builder,
) error {
	// if observed generation is out of sync, treat the issued certificate as Pending (spec has been modified)
	if issuedCertificate.Status.ObservedGeneration != issuedCertificate.Generation {
		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_PENDING
	}

	// reset & update status
	issuedCertificate.Status.ObservedGeneration = issuedCertificate.Generation
	issuedCertificate.Status.Error = ""

	// state-machine style processor
	switch issuedCertificate.Status.State {
	case v1alpha2.IssuedCertificateStatus_FINISHED:
		// ensure issued cert secret exists, nothing to do for this issued certificate
		if issuedCertificateSecret, err := inputSecrets.Find(issuedCertificate.Spec.IssuedCertificateSecret); err == nil {
			// add secret output to prevent it from being GC'ed
			outputs.AddSecrets(issuedCertificateSecret)
			return nil
		}
		// otherwise, restart the workflow from PENDING
		fallthrough
	case v1alpha2.IssuedCertificateStatus_FAILED:
		// restart the workflow from PENDING
		fallthrough
	case v1alpha2.IssuedCertificateStatus_PENDING:
		// create a new private key
		privateKey, err := utils.GeneratePrivateKey()
		if err != nil {
			return err
		}
		outputs.AddSecrets(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuedCertificate.Name,
				Namespace: issuedCertificate.Namespace,
				Labels:    agentLabels,
			},
			Data: map[string][]byte{privateKeySecretKey: privateKey},
			Type: privateKeySecretType,
		})

		// create certificate request for private key
		csrBytes, err := utils.GenerateCertificateSigningRequest(
			issuedCertificate.Spec.Hosts,
			issuedCertificate.Spec.Org,
			privateKey,
		)
		if err != nil {
			return err
		}
		certificateRequest := &v1alpha2.CertificateRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuedCertificate.Name,
				Namespace: issuedCertificate.Namespace,
				Labels:    agentLabels,
			},
			Spec: v1alpha2.CertificateRequestSpec{
				CertificateSigningRequest: csrBytes,
			},
		}
		outputs.AddCertificateRequests(certificateRequest)

		// set status to REQUESTED
		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_REQUESTED
	case v1alpha2.IssuedCertificateStatus_REQUESTED:

		// retrieve private key
		privateKeySecret, err := inputSecrets.Find(issuedCertificate)
		if err != nil {
			return err
		}
		privateKey := privateKeySecret.Data[privateKeySecretKey]

		if len(privateKey) == 0 {
			return eris.Errorf("invalid private key found, no data provided")
		}

		// retrieve signed certificate
		certificateRequest, err := inputCertificateRequests.Find(issuedCertificate)
		if err != nil {
			return err
		}

		switch certificateRequest.Status.State {
		case v1alpha2.CertificateRequestStatus_PENDING:
			contextutils.LoggerFrom(r.ctx).Infof("waiting for certificate request %v to be signed by Issuer", sets.Key(certificateRequest))

			// add secret and certrequest to output to prevent them from being GC'ed
			outputs.AddSecrets(privateKeySecret)
			outputs.AddCertificateRequests(certificateRequest)

			// if the certificate signing request has not been
			// fulfilled, return and wait for the next reconcile
			return nil
		case v1alpha2.CertificateRequestStatus_FAILED:
			return eris.Errorf("certificate request %v failed to be signed by Issuer: %v", sets.Key(certificateRequest), certificateRequest.Status.Error)
		}

		signedCert := certificateRequest.Status.SignedCertificate
		signingRootCA := certificateRequest.Status.SigningRootCa

		issuedCertificateData := secrets.IntermediateCAData{
			RootCAData: secrets.RootCAData{
				RootCert: signingRootCA,
			},
			CertChain:    utils.AppendRootCerts(signedCert, signingRootCA),
			CaCert:       signedCert,
			CaPrivateKey: privateKey,
		}

		// finally, create the secret
		issuedCertificateSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuedCertificate.Spec.IssuedCertificateSecret.Name,
				Namespace: issuedCertificate.Spec.IssuedCertificateSecret.Namespace,
				Labels:    agentLabels,
			},
			Data: issuedCertificateData.ToSecretData(),
			Type: issuedCertificateSecretType,
		}
		outputs.AddSecrets(issuedCertificateSecret)

		// mark issued certificate as ISSUED
		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_ISSUED
	case v1alpha2.IssuedCertificateStatus_ISSUED:
		// ensure issued cert secret exists, if not, return an error (restart the workflow)
		if issuedCertificateSecret, err := inputSecrets.Find(issuedCertificate.Spec.IssuedCertificateSecret); err != nil {
			return err
		} else {
			// add secret output to prevent it from being GC'ed
			outputs.AddSecrets(issuedCertificateSecret)
		}

		// now we must bounce all the pods
		if err := r.bouncePods(issuedCertificate.Spec.PodBounceDirective, inputPods); err != nil {
			return eris.Wrap(err, "bouncing pods")
		}

		// mark issued certificate as finished
		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_FINISHED
	default:
		return eris.Errorf("unknown issued certificate state: %v", issuedCertificate.Status.State)
	}

	return nil
}

// bounce (delete) the listed pods
func (r *certAgentReconciler) bouncePods(podBounceDirectiveRef *v1.ObjectRef, allPods corev1sets.PodSet) error {
	if podBounceDirectiveRef == nil {
		return nil
	}
	podBounceDirective, err := v1alpha2.NewPodBounceDirectiveClient(r.localClient).GetPodBounceDirective(r.ctx, ezkube.MakeClientObjectKey(podBounceDirectiveRef))
	if err != nil {
		return eris.Wrap(err, "failed to find specified pod bounce directive")
	}

	podsToDelete := allPods.List(func(pod *corev1.Pod) bool {
		for _, selector := range podBounceDirective.Spec.PodsToBounce {
			if selector.Namespace == pod.Namespace &&
				labels.SelectorFromSet(selector.Labels).Matches(labels.Set(pod.Labels)) {
				return false
			}
		}
		return true
	})

	podClient := corev1client.NewPodClient(r.localClient)
	var errs error
	for _, pod := range podsToDelete {
		contextutils.LoggerFrom(r.ctx).Debugf("deleting pod %v", sets.Key(pod))
		if err := podClient.DeletePod(r.ctx, ezkube.MakeClientObjectKey(pod)); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}
