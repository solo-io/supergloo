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
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

func collectLogs(ctx context.Context, opts *DebugReportOpts, dir string, config utils.MeshctlConfig) error {
	logsDir := dir + "/logs"
	err := opts.fs.MkdirAll(logsDir, 0755)
	if err != nil {
		return err
	}

	var responses []*debugutils.LogsResponse
	for name, cluster := range config.Clusters {
		clusterLogsDir := logsDir + "/" + name
		err = opts.fs.MkdirAll(clusterLogsDir, 0755)
		if err != nil {
			return err
		}

		cfg, err := kubeconfig.GetRestConfigWithContext(cluster.KubeConfig, cluster.KubeContext, "")
		if err != nil {
			return errors.Wrapf(err, "getting kube config")
		}
		kubeClient, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return errors.Wrapf(err, "getting kube clientset")
		}
		logCollector := debugutils.NewLogCollector(debugutils.NewLogRequestBuilder(kubeClient.CoreV1(),
			debugutils.NewLabelPodFinder(kubeClient)))
		if err != nil {
			return err
		}

		unstructuredPods, err := collectUnstructuredPods(ctx, opts, kubeClient.CoreV1(), config.IsMgmtCluster(name))
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
		err = eg.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

func collectUnstructuredPods(ctx context.Context, opts *DebugReportOpts, clientset corev1client.CoreV1Interface, isMgmtCluster bool) ([]*unstructured.Unstructured, error) {
	var unstructuredPods []*unstructured.Unstructured
	var err error
	if isMgmtCluster {
		unstructuredPods, err = getUnstructuredPods(ctx, clientset, opts.namespace,
			"app in (gloo-mesh, enterprise-networking)")
		if err != nil {
			return unstructuredPods, err
		}
	} else {
		glooMeshLogRequests, err := getUnstructuredPods(ctx, clientset, opts.namespace,
			"app in (enterprise-agent)")
		if err != nil {
			return unstructuredPods, err
		}
		unstructuredPods = append(unstructuredPods, glooMeshLogRequests...)
		istioLogRequests, err := getUnstructuredPods(ctx, clientset, "istio-system",
			"app in (istiod)")
		if err != nil {
			return unstructuredPods, err
		}
		unstructuredPods = append(unstructuredPods, istioLogRequests...)
	}
	return unstructuredPods, err
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
