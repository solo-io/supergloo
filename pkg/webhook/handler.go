package webhook

import (
	"context"
	"strings"

	"github.com/solo-io/supergloo/pkg/webhook/clients"

	"github.com/solo-io/go-utils/errors"

	"github.com/solo-io/go-utils/contextutils"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handlePodCreation(w http.ResponseWriter, r *http.Request) {
	logger := contextutils.LoggerFrom(r.Context())
	logger.Infof("received pod creation request: %v", r)

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	// Verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Warnf("contentType=%s, expecting application/json", contentType)
		return
	}

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err != nil {
			logger.Errorf("failed to read request body. Error: %s", err.Error())
			return
		} else {
			body = data
		}
	}

	deserializer := clients.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		responseAdmissionReview.Response = toErrorResponse(r.Context(), fmt.Sprintf("failed to decode request body. Error: %s", err.Error()))
	} else {
		response, err := admit(r.Context(), requestedAdmissionReview)
		if err != nil {
			response = toErrorResponse(r.Context(), err.Error())
		}
		responseAdmissionReview.Response = response
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	logger.Debugf("sending AdmissionReview response: %v", responseAdmissionReview.Response)
	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		logger.Errorf("failed to marshal AdmissionReview response. Error: %s", err.Error())
	}
	if _, err := w.Write(respBytes); err != nil {
		logger.Errorf("failed to write AdmissionReview response. Error: %s", err.Error())
	}
}

func admit(ctx context.Context, ar v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
	logger := contextutils.LoggerFrom(ctx)

	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		return nil, errors.Errorf("expected resource to be: %s, but found: %v", podResource, ar.Request.Resource)
	}

	pod := corev1.Pod{}
	if _, _, err := clients.Codecs.UniversalDeserializer().Decode(ar.Request.Object.Raw, nil, &pod); err != nil {
		return nil, errors.Wrapf(err, "failed to deserialize raw pod resource")
	}
	logger.Infof("evaluating pod %s.%s for sidecar auto-injection", pod.Namespace, pod.Name)

	// Check if pod need to be injected with sidecar and, if so, generate the correspondent patch
	patchRequired, patch, err := GetInjectionHandler().GetSidecarPatch(ctx, &pod)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create sidecar patch")
	}

	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if patchRequired {
		pt := v1beta1.PatchTypeJSONPatch
		reviewResponse.PatchType = &pt
		reviewResponse.Patch = patch
		reviewResponse.Result = &metav1.Status{
			Status:  "Success",
			Message: strings.TrimSpace("successfully injected pod with AWS App Mesh Envoy sidecar")}
	}

	return &reviewResponse, nil
}

func toErrorResponse(ctx context.Context, errMsg string) *v1beta1.AdmissionResponse {
	contextutils.LoggerFrom(ctx).Error(errMsg)
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			// These error messages will be returned by the k8s api server, potentially on operations where the origin
			// of the error might not be immediately evident. The prefix makes it clear that is was us.
			Message: "supergloo: " + errMsg,
			Status:  "Failure",
		},
	}
}
