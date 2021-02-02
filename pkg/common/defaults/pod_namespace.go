package defaults

import "os"

const (
	PodNamespaceEnv     = "POD_NAMESPACE"
	AgentClusterEnv     = "AGENT_CLUSTER"
	DefaultPodNamespace = "gloo-mesh"
)

func GetPodNamespace() string {
	if podNamespace := os.Getenv(PodNamespaceEnv); podNamespace != "" {
		return podNamespace
	}
	return DefaultPodNamespace
}

func GetAgentCluster() string {
	return os.Getenv(AgentClusterEnv)
}
