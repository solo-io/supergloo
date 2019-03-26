package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/solo-io/go-utils/contextutils"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const HandlePattern = "/pods/inject-sidecar"

type AdmissionWebhookServer struct {
	*http.Server
}

func NewSidecarInjectionWebhook(ctx context.Context) *AdmissionWebhookServer {
	logger := contextutils.LoggerFrom(ctx)

	// Set global client set
	if err := initClientSet(ctx); err != nil {
		logger.Fatalf("failed to create webhook client set: %v", err)
	}

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

	deserializer := Codecs.UniversalDeserializer()
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
