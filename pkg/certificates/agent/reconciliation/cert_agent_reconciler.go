package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	corev1client "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/output/certagent"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/utils"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/output/errhandlers"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
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
			inputSnap.ConfigMaps(),
			inputSnap.CertificateRequests(),
			inputSnap.PodBounceDirectives(),
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
	if err := inputSnap.SyncStatuses(r.ctx, r.localClient, input.SyncStatusOptions{
		IssuedCertificate:  true,
		PodBounceDirective: true,
	}); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

func (r *certAgentReconciler) reconcileIssuedCertificate(
	issuedCertificate *v1alpha2.IssuedCertificate,
	inputSecrets corev1sets.SecretSet,
	inputPods corev1sets.PodSet,
	inputConfigMaps corev1sets.ConfigMapSet,
	inputCertificateRequests v1alpha2sets.CertificateRequestSet,
	podBounceDirectives v1alpha2sets.PodBounceDirectiveSet,
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

		// see if we need to bounce pods
		if issuedCertificate.Spec.PodBounceDirective != nil {
			podBounceDirective, err := podBounceDirectives.Find(issuedCertificate.Spec.PodBounceDirective)
			if err != nil {
				return eris.Wrap(err, "failed to find specified pod bounce directive")
			}

			// try to bounce the pods and see if we need to wait
			waitForConditions, err := r.bouncePods(podBounceDirective, inputPods, inputConfigMaps, inputSecrets)
			if err != nil {
				return eris.Wrap(err, "bouncing pods")
			}

			if waitForConditions {
				// return here without updating the status of the issued Certificate; we want to retry bouncing the pods
				// when replacements are live and data plane certs are distributed
				return nil
			}
		}

		// mark issued certificate as finished
		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_FINISHED
	default:
		return eris.Errorf("unknown issued certificate state: %v", issuedCertificate.Status.State)
	}

	return nil
}

// bounce (delete) the listed pods
// returns true if we need to wait before proceeding to process the podBounceDirective.
// we must wait for the following conditions:
// 1. istiod control plane has come back online after it has been restarted
// 2. istio's root cert has been propagated to all istio-controlled namespaces for consumption by the data plane.
// this will cause the reconcile to end early and persist the IssuedCertificate in the Issued state
func (r *certAgentReconciler) bouncePods(podBounceDirective *v1alpha2.PodBounceDirective, allPods corev1sets.PodSet, allConfigMaps corev1sets.ConfigMapSet, allSecrets corev1sets.SecretSet) (bool, error) {

	// create a client here to call for deletions
	podClient := corev1client.NewPodClient(r.localClient)

	var errs error

	// collect the pods we want to delete in the order they're specified in the directive
	// it is important Istiod is restarted before any of the other pods
	for i, selector := range podBounceDirective.Spec.PodsToBounce {
		if len(podBounceDirective.Status.PodsBounced) > i {
			// the set of pods for this selector was already bounced, so we instead ensure that
			// the minimum number of replacement pods are ready before moving on with the deletion
			podsBounced := podBounceDirective.Status.PodsBounced[i]

			// if all required replicas are not ready, return true to indicate we should halt processing
			// of the directive here in order to wait for a future update to the Pods in the input snapshot.
			if !replacementsReady(allPods, selector, podsBounced.BouncedPods) {

				contextutils.LoggerFrom(r.ctx).Debugf("podBounceDirective %v: waiting for ready pods for selector %v", sets.Key(podBounceDirective), selector)

				time.Sleep(time.Second)

				// wait for all replicas of these pods to be ready before proceeding to the next selector
				// used to ensure gateway restart does not preempt control plane restart
				return true, errs
			}

			// skip deletion, these pods were already bounced
			continue
		}

		if selector.RootCertSync != nil {
			configMap, err := allConfigMaps.Find(selector.RootCertSync.ConfigMapRef)
			if err != nil && errors.IsNotFound(err) {
				// ConfigMap isn't found; let's wait for it to be added by Istio
				time.Sleep(time.Second)

				return true, nil
			} else if err != nil {
				return true, err
			}

			secret, err := allSecrets.Find(selector.RootCertSync.SecretRef)
			if err != nil {
				return true, err
			}

			if configMap.Data[selector.RootCertSync.ConfigMapKey] != string(secret.Data[selector.RootCertSync.SecretKey]) {
				// the configmap's public key doesn't match the root cert CA's
				// sleep to allow time for the cert to be distributed and retry
				time.Sleep(time.Second)

				return true, nil
			}
		}

		podsToDelete := allPods.List(func(pod *corev1.Pod) bool {
			return !isPodSelected(pod, selector)
		})

		var bouncedPods []string
		for _, pod := range podsToDelete {
			contextutils.LoggerFrom(r.ctx).Debugf("deleting pod %v", sets.Key(pod))
			if err := podClient.DeletePod(r.ctx, ezkube.MakeClientObjectKey(pod)); err != nil {
				errs = multierror.Append(errs, err)
			}
			bouncedPods = append(bouncedPods, pod.Name)
		}

		// update the status to show we've bounced this selector already
		podBounceDirective.Status.PodsBounced = append(
			podBounceDirective.Status.PodsBounced,
			&v1alpha2.PodBounceDirectiveStatus_BouncedPodSet{BouncedPods: bouncedPods},
		)

		if selector.WaitForReplicas > 0 {
			// if we just deleted pods for a selector that wants us to wait, we should return here
			// and signal a wait
			return true, errs
		}
	}

	return false, errs
}

// indicates whether replacements for the deleted pods to be ready
func replacementsReady(
	currentPods corev1sets.PodSet,
	podSelector *v1alpha2.PodBounceDirectiveSpec_PodSelector,
	deletedPodNames []string,
) bool {
	if podSelector.WaitForReplicas == 0 {
		return true
	}

	// get the list of pods that are in ready condition
	currentReadyPods := currentPods.List(func(pod *corev1.Pod) bool {
		for _, deletedPodName := range deletedPodNames {
			if deletedPodName == pod.Name {
				// exclude pods that have been recently deleted, but still appear in the snapshot
				return true
			}
		}
		// exclude pods that are not ready && selected
		return !(isPodSelected(pod, podSelector) && podutil.IsPodReady(pod))
	})

	// ensure we have the right number of ready replicas
	return len(currentReadyPods) >= int(podSelector.WaitForReplicas)
}

func isPodSelected(pod *corev1.Pod, podSelector *v1alpha2.PodBounceDirectiveSpec_PodSelector) bool {
	return podSelector.Namespace == pod.Namespace &&
		labels.SelectorFromSet(podSelector.Labels).Matches(labels.Set(pod.Labels))
}
