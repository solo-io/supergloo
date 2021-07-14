package checks

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/rotisserie/eris"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mgmtDeployName = "enterprise-networking"
	clusterRegDoc  = "https://docs.solo.io/gloo-mesh/latest/setup/cluster_registration/enterprise_cluster_registration/"
)

const (
	// NOTE: these must correspond 1:1 with the metric names defined here:
	// https://github.com/solo-io/skv2-enterprise/blob/master/relay/pkg/pull/pull_server.go#L23
	// https://github.com/solo-io/skv2-enterprise/blob/master/relay/pkg/push/push_server.go#L31
	metricClusterLabel             = "cluster"
	connectedPullClientsMetricName = "relay_pull_clients_connected"
	connectedPushClientsMetricName = "relay_push_clients_connected"
)

type enterpriseRegistrationCheck struct {
}

func NewEnterpriseRegistrationCheck() *enterpriseRegistrationCheck {
	return &enterpriseRegistrationCheck{}
}

func (d *enterpriseRegistrationCheck) GetDescription() string {
	return "Gloo Mesh agents are connected for each registered KubernetesCluster."
}

type connectionStatus struct {
	cluster       string
	registered    bool
	agentsPulling int
	agentsPushing int
}

func isEnterpriseVersion(ctx context.Context, c client.Client, installNamespace string) (bool, error) {
	_, err := v1.NewDeploymentClient(c).GetDeployment(ctx, client.ObjectKey{
		Namespace: installNamespace,
		Name:      mgmtDeployName,
	})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *enterpriseRegistrationCheck) Run(ctx context.Context, checkCtx CheckContext) *Failure {
	shouldRunCheck, err := isEnterpriseVersion(ctx, checkCtx.Client(), checkCtx.Environment().Namespace)
	if err != nil {
		return &Failure{
			Errors: []error{err},
		}
	}

	if !shouldRunCheck {
		contextutils.LoggerFrom(ctx).Debugf("skipping relay connectivity check, enterprise not detected")
		return nil
	}

	if checkCtx.Environment().AdminPort == 0 {
		contextutils.LoggerFrom(ctx).Debugf("skipping relay connectivity check, remote port set to 0")
		return nil
	}

	installNamespace := checkCtx.Environment().Namespace
	// get registered clusters
	registeredClusters, err := v1alpha1.NewKubernetesClusterClient(checkCtx.Client()).ListKubernetesCluster(ctx, client.InNamespace(installNamespace))
	if err != nil {
		return &Failure{
			Errors: []error{err},
		}
	}
	failure := new(Failure)
	if len(registeredClusters.Items) == 0 {
		failure.AddHint("You don't have any registered clusters. you may want to create a KubernetesCluster CR.", clusterRegDoc)
	}

	// get connected clusters
	connectedPullAgents, connectedPushAgents, err, hint := d.getConnectedAgents(ctx, checkCtx)
	if err != nil {
		failure.AddHint(hint, clusterRegDoc)
		failure.AddError(err)
		return failure
	}

	clusterStatuses := calculateClusterStatuses(registeredClusters, connectedPullAgents, connectedPushAgents)

	printClusterStatuses(clusterStatuses)

	// consider each cluster that is in a partially registered state to be an error
	for _, status := range clusterStatuses {
		switch {
		// cluster registered, agents are not pulling or pushing
		case status.registered && ((status.agentsPulling < 1) || (status.agentsPushing < 1)):
			failure.AddError(eris.Errorf("cluster %v registered but agent is not connected (pull: %v, push: %v)", status.cluster, status.agentsPulling, status.agentsPushing))
			failure.AddHint("check the logs of the agent on "+status.cluster+" and investigate whether/why the gRPC connection failed from the agent to the mgmt server. "+
				"Additionally, if the pushing value zero and pulling is non-zero, that usually indicates a connected dashboard and an unconnected agent.", clusterRegDoc)
		// cluster not registered, agents are pushing. Cannot use positive pull status as agent evidence, since it may also be a dashboard connection.
		case !status.registered && status.agentsPushing > 0:
			failure.AddError(eris.Errorf("cluster %v is not currently registered but agent is connected (pull: %v, push: %v)", status.cluster, status.agentsPulling, status.agentsPushing))
			failure.AddHint("create a corresponding KubernetesCluster CR to register the "+status.cluster+".", clusterRegDoc)
		}
	}

	return failure
}

func printClusterStatuses(clusterStatuses []connectionStatus) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Cluster", "Registered", "Dashboards and Agents Pulling", " Agents Pushing"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, status := range clusterStatuses {
		table.Append([]string{
			status.cluster,
			fmt.Sprintf("%v", status.registered),
			fmt.Sprintf("%v", status.agentsPulling),
			fmt.Sprintf("%v", status.agentsPushing),
		})
	}
	table.Render()

}

func calculateClusterStatuses(
	registeredClusters *v1alpha1.KubernetesClusterList,
	connectedPullAgents, connectedPushAgents map[string]int,
) []connectionStatus {

	clusterStatuses := map[string]connectionStatus{}

	for _, registeredCluster := range registeredClusters.Items {
		cluster := registeredCluster.Name
		clusterStatuses[cluster] = connectionStatus{
			cluster:       cluster,
			registered:    true,
			agentsPulling: connectedPullAgents[cluster],
			agentsPushing: connectedPushAgents[cluster],
		}
	}
	// find agents who are connected/disconnected for unregistered clusters
	for cluster, connected := range connectedPullAgents {
		if _, ok := clusterStatuses[cluster]; !ok {
			clusterStatuses[cluster] = connectionStatus{
				cluster:       cluster,
				registered:    false,
				agentsPulling: connected,
				agentsPushing: connectedPushAgents[cluster],
			}
		}
	}
	for cluster, connected := range connectedPushAgents {
		if _, ok := clusterStatuses[cluster]; !ok {
			clusterStatuses[cluster] = connectionStatus{
				cluster:       cluster,
				registered:    false,
				agentsPulling: connectedPullAgents[cluster],
				agentsPushing: connected,
			}
		}
	}
	var sortedStatuses []connectionStatus
	for _, status := range clusterStatuses {
		sortedStatuses = append(sortedStatuses, status)
	}
	sort.SliceStable(sortedStatuses, func(i, j int) bool {
		return sortedStatuses[i].cluster < sortedStatuses[j].cluster
	})
	return sortedStatuses
}

func (d *enterpriseRegistrationCheck) getConnectedAgents(ctx context.Context, checkCtx CheckContext) (map[string]int, map[string]int, error, string) {

	var parsedMetrics map[string]*dto.MetricFamily
	err, hint := checkCtx.AccessAdminPort(ctx, mgmtDeployName, func(ctx context.Context, adminUrl *url.URL) (error, string) {
		metricsUrl := adminUrl.ResolveReference(&url.URL{Path: "/metrics"}).String()
		resp, err := http.DefaultClient.Get(metricsUrl)
		if err != nil {
			return err, fmt.Sprintf("try verifying that the mgmt pods are listening on port %v", checkCtx.Environment().AdminPort)
		}
		defer resp.Body.Close()

		parsedMetrics, err = (&expfmt.TextParser{}).TextToMetricFamilies(resp.Body)
		if err != nil {
			return err, fmt.Sprintf("try verifying that the mgmt pods serving metrics at %v", metricsUrl)
		}
		return nil, ""
	})
	if err != nil {
		return nil, nil, err, hint
	}

	connectedPullClientsMetric, ok := parsedMetrics[connectedPullClientsMetricName]
	if !ok {
		return nil, nil, eris.Errorf("expected metric %v not present.", connectedPullClientsMetricName), "Try verifying that at least one agent has connected to the management plane."
	}

	// map of cluster to connection status;
	// false for disconnected (gauge value 0)
	// true for connected (gauge value 1)
	pullClientsConnected, err, hint := getClientClustersConnected(connectedPullClientsMetric)
	if err != nil {
		return nil, nil, err, hint
	}

	connectedPushClientsMetric, ok := parsedMetrics[connectedPushClientsMetricName]
	if !ok {
		return nil, nil, eris.Errorf("expected metric %v not present", connectedPushClientsMetricName), "Try verifying that at least one agent has connected to the management plane."
	}

	pushClientsConnected, err, reason := getClientClustersConnected(connectedPushClientsMetric)
	if err != nil {
		return nil, nil, err, reason
	}

	return pullClientsConnected, pushClientsConnected, nil, ""
}

func getClientClustersConnected(clientConnectionMetrics *dto.MetricFamily) (map[string]int, error, string) {
	clustersConnected := map[string]int{}
	for _, metric := range clientConnectionMetrics.GetMetric() {
		var cluster string
		for _, pair := range metric.GetLabel() {
			if pair.GetName() == metricClusterLabel {
				cluster = pair.GetValue()
			}
		}
		if cluster == "" {
			return nil, eris.Errorf("parsed metrics missing cluster label %v", clientConnectionMetrics), "ensure your version of `meshctl` matches the installed version of Gloo Mesh Enterprise."
		}
		// gauge value can either be 0 for disconnected, 1 for connected
		clustersConnected[cluster] = int(metric.GetGauge().GetValue())
	}
	return clustersConnected, nil, ""
}
