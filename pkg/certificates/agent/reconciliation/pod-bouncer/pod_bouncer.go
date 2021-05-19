package pod_bouncer

import (
	"bytes"
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	corev1client "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

//go:generate mockgen -source ./pod_bouncer.go -destination mocks/pod_bouncer.go

type RootCertMatcher interface {
	MatchesRootCert(
		ctx context.Context,
		rootCert []byte,
		selector *certificatesv1.PodBounceDirectiveSpec_PodSelector,
		allSecrets corev1sets.SecretSet,
	) (matches bool, err error)
}

// bounce (delete) the listed pods
// returns true if we need to wait before proceeding to process the podBounceDirective.
// we must wait for the following conditions:
// 1. istiod control plane has come back online after it has been restarted
// 2. istio's root cert has been propagated to all istio-controlled namespaces for consumption by the data plane.
// this should cause the reconcile to end early and persist the IssuedCertificate in the Issued state
type PodBouncer interface {
	BouncePods(
		ctx context.Context,
		podBounceDirective *certificatesv1.PodBounceDirective,
		pods corev1sets.PodSet,
		configMaps corev1sets.ConfigMapSet,
		secrets corev1sets.SecretSet,
	) (bool, error)
}

func NewPodBouncer(
	podClient corev1client.PodClient,
	rootCertMatcher RootCertMatcher,
) PodBouncer {
	return &podBouncer{
		podClient:       podClient,
		rootCertMatcher: rootCertMatcher,
	}
}

type podBouncer struct {
	podClient       corev1client.PodClient
	rootCertMatcher RootCertMatcher
}

// TODO: Figure out how to bounce pods without any downtime
func (p *podBouncer) BouncePods(
	ctx context.Context,
	podBounceDirective *certificatesv1.PodBounceDirective,
	pods corev1sets.PodSet,
	configMaps corev1sets.ConfigMapSet,
	secrets corev1sets.SecretSet,
) (bool, error) {

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
			if !replacementsReady(pods, selector, podsBounced.BouncedPods) {

				contextutils.LoggerFrom(ctx).Debugf("podBounceDirective %v: waiting for ready pods for selector %v", sets.Key(podBounceDirective), selector)

				time.Sleep(time.Second)

				// wait for all replicas of these pods to be ready before proceeding to the next selector
				// used to ensure gateway restart does not preempt control plane restart
				return true, errs
			}

			// skip deletion, these pods were already bounced
			continue
		}

		if selector.RootCertSync != nil {
			configMap, err := configMaps.Find(selector.RootCertSync.ConfigMapRef)
			if err != nil && errors.IsNotFound(err) {
				// ConfigMap isn't found; let's wait for it to be added by Istio
				contextutils.LoggerFrom(ctx).Debugf("podBounceDirective %v: waiting for %v configmap creation for selector %v", sets.Key(podBounceDirective), selector.RootCertSync.ConfigMapRef.Name, selector)

				time.Sleep(time.Second)

				return true, nil
			} else if err != nil {
				return true, err
			}

			matches, err := p.rootCertMatcher.MatchesRootCert(
				ctx,
				[]byte(configMap.Data[selector.RootCertSync.ConfigMapKey]),
				selector,
				secrets,
			)
			if err != nil {
				return true, err
			}

			if !matches {
				// the configmap's public key doesn't match the root cert CA's
				// sleep to allow time for the cert to be distributed and retry
				contextutils.LoggerFrom(ctx).Debugf("podBounceDirective %v: waiting for %v configmap update for selector %v", sets.Key(podBounceDirective), selector.RootCertSync.ConfigMapRef.Name, selector)

				time.Sleep(time.Second)

				return true, nil
			}
		}

		podsToDelete := pods.List(func(pod *corev1.Pod) bool {
			return !isPodSelected(pod, selector)
		})

		var bouncedPods []string
		for _, pod := range podsToDelete {
			contextutils.LoggerFrom(ctx).Debugf("deleting pod %v", sets.Key(pod))
			if err := p.podClient.DeletePod(ctx, ezkube.MakeClientObjectKey(pod)); err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
			bouncedPods = append(bouncedPods, pod.Name)
		}

		// update the status to show we've bounced this selector already
		podBounceDirective.Status.PodsBounced = append(
			podBounceDirective.Status.PodsBounced,
			&certificatesv1.PodBounceDirectiveStatus_BouncedPodSet{BouncedPods: bouncedPods},
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
	podSelector *certificatesv1.PodBounceDirectiveSpec_PodSelector,
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

func isPodSelected(pod *corev1.Pod, podSelector *certificatesv1.PodBounceDirectiveSpec_PodSelector) bool {
	return podSelector.Namespace == pod.Namespace &&
		labels.SelectorFromSet(podSelector.Labels).Matches(labels.Set(pod.Labels))
}

func NewSecretRootCertMatcher() RootCertMatcher {
	return &secretRootCertMatcher{}
}

type secretRootCertMatcher struct {
}

func (s *secretRootCertMatcher) MatchesRootCert(
	ctx context.Context,
	rootCert []byte,
	selector *certificatesv1.PodBounceDirectiveSpec_PodSelector,
	allSecrets corev1sets.SecretSet,
) (matches bool, err error) {
	secret, err := allSecrets.Find(selector.RootCertSync.SecretRef)
	if err != nil {
		return false, err
	}

	return bytes.Equal(secret.Data[selector.RootCertSync.SecretKey], rootCert), nil
}
