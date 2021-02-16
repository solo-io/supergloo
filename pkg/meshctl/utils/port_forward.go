package utils

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Open a port-forward against the specified deployment. Returns the stop channel and the local port.
// If localPort is unspecified, a free port will be chosen at random.
// Close the port forward by closing the returned channel.
func PortForwardFromDeployment(
	ctx context.Context,
	kubeConfig string,
	kubeContext string,
	deployName string,
	deployNamespace string,
	localPort string,
	remotePort string,
	logger *logrus.Logger,
) (chan struct{}, string, error) {
	podName, err := getPodForDeployment(ctx, kubeConfig, kubeContext, deployName, deployNamespace)
	if err != nil {
		return nil, "", err
	}
	return PortForwardFromPod(kubeConfig, kubeContext, podName, deployNamespace, localPort, remotePort, logger)
}

// Open a port forward against the specified pod. Returns the stop channel and the local port.
// If localPort is unspecified, a free port will be chosen at random.
// Close the port forward by closing the returned channel.
func PortForwardFromPod(
	kubeConfig string,
	kubeContext string,
	podName string,
	podNamespace string,
	localPort string,
	remotePort string,
	logger *logrus.Logger,
) (chan struct{}, string, error) {
	// select random open local port if unspecified
	if localPort == "" {
		freePort, err := cliutils.GetFreePort()
		if err != nil {
			return nil, "", err
		}
		localPort = strconv.Itoa(freePort)
		logger.Debugf("forwarding port %s of pod %s in namespace %s to local port %s", remotePort, podName, podNamespace, localPort)
	}

	config, err := kubeconfig.GetRestConfigWithContext(kubeConfig, kubeContext, "")
	if err != nil {
		return nil, "", err
	}

	// the following code is based on this reference, https://github.com/kubernetes/client-go/issues/51
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, "", err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", podNamespace, podName)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%s:%s", localPort, remotePort)}, stopChan, readyChan, out, errOut)
	if err != nil {
		return nil, "", err
	}

	if len(errOut.String()) != 0 {
		return nil, "", eris.New(errOut.String())
	} else if len(out.String()) != 0 {
		logger.Debug(out.String())
	}

	go func() {
		if err = forwarder.ForwardPorts(); err != nil { // Locks until stopChan is closed.
			logger.Errorf("%v", err)
		}
	}()

	// block until port forward is ready
	select {
	case <-readyChan:
		break
	}

	return stopChan, localPort, nil
}

// select a pod backing a deployment
func getPodForDeployment(
	ctx context.Context,
	kubeConfig string,
	kubeContext string,
	deploymentName string,
	deploymentNamespace string,
) (string, error) {
	config, err := kubeconfig.GetRestConfigWithContext(kubeConfig, kubeContext, "")
	if err != nil {
		return "", err
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	deployment, err := kubeClient.AppsV1().Deployments(deploymentNamespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	matchLabels := deployment.Spec.Selector.MatchLabels
	listOptions := (&client.ListOptions{LabelSelector: labels.SelectorFromSet(matchLabels)}).AsListOptions()

	podList, err := kubeClient.CoreV1().Pods(deploymentNamespace).List(ctx, *listOptions)
	if err != nil {
		return "", err
	}

	// select the first backing pod
	for _, pod := range podList.Items {
		return pod.Name, nil
	}

	return "", eris.New("no pods found")
}
