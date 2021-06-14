package report

import (
	"bufio"
	"context"
	"io"
	"strings"

	"encoding/json"

	errors "github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/go-utils/errgroup"
	"github.com/solo-io/k8s-utils/debugutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

func collectLogs(ctx context.Context, opts *DebugReportOpts, clusterLogsDir string, cluster utils.MeshctlCluster, isMgmt bool) error {
	var responses []*debugutils.LogsResponse
	kubeClient, err := utils.BuildClientset(cluster.KubeConfig, cluster.KubeContext)
	if err != nil {
		return errors.Wrapf(err, "getting kube clientset")
	}
	logCollector := debugutils.NewLogCollector(debugutils.NewLogRequestBuilder(kubeClient.CoreV1(),
		debugutils.NewLabelPodFinder(kubeClient)))

	var labelSelector string
	if isMgmt {
		labelSelector = "app in (gloo-mesh, enterprise-networking, discovery, networking, cert-agent)"
	} else {
		labelSelector = "app in (enterprise-agent)"
	}
	unstructuredPods, err := getUnstructuredPods(ctx, kubeClient.CoreV1(), opts.namespace,
		labelSelector)
	if err != nil {
		return err
	}
	logRequests, err := logCollector.GetLogRequests(ctx, unstructuredPods)
	if err != nil {
		return err
	}
	clusterLogResponses, err := logCollector.LogRequestBuilder.StreamLogs(ctx, logRequests)
	if err != nil {
		return err
	}
	responses = append(responses, clusterLogResponses...)
	// Write logs into a directory
	eg := errgroup.Group{}
	for _, response := range responses {
		response := response
		eg.Go(func() error {
			defer response.Response.Close()
			logs := readLogs(response.Response)
			if logs.Len() > 0 {
				err = debugutils.NewFileStorageClient(opts.fs).Save(clusterLogsDir, &debugutils.StorageObject{
					Resource: strings.NewReader(logs.String()),
					Name:     response.ResourceId(),
				})
			}
			return nil
		})
	}
	return eg.Wait()
}

func getUnstructuredPods(ctx context.Context, clientset corev1client.CoreV1Interface, ns, labelSelector string) ([]*unstructured.Unstructured, error) {
	pods, err := clientset.Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	resources, err := debugutils.ConvertPodsToUnstructured(pods)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func readLogs(r io.ReadCloser) strings.Builder {
	scanner := bufio.NewScanner(r)
	logs := strings.Builder{}
	for scanner.Scan() {
		line := scanner.Text()
		start := strings.Index(line, "{")
		if start == -1 {
			// Not a json formatted log
			logs.WriteString(line + "\n")
			continue
		}
		in := []byte(line[start:])
		var raw map[string]interface{}
		if err := json.Unmarshal(in, &raw); err != nil {
			continue
		}
		if raw["level"] == "warn" || raw["level"] == "error" || raw["level"] == "fatal" {
			logs.WriteString(line + "\n")
		}
	}
	return logs
}
