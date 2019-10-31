package utils

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

func DeploymentsByCluster(deployments kubernetes.DeploymentList) map[string]kubernetes.DeploymentList {
	byCluster := make(map[string]kubernetes.DeploymentList)
	for _, d := range deployments {
		byCluster[d.ClusterName] = append(byCluster[d.ClusterName], d)
	}
	return byCluster
}

func UpstreamsByCluster(upstreams gloov1.UpstreamList) map[string]gloov1.UpstreamList {
	byCluster := make(map[string]gloov1.UpstreamList)
	for _, u := range upstreams {
		byCluster[u.Metadata.Cluster] = append(byCluster[u.Metadata.Cluster], u)
	}
	return byCluster
}
