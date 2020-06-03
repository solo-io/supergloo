package metadata

import "fmt"

const (
	EksPrefix = "eks"
)

func BuildEksKubernetesClusterName(clusterName, region string) string {
	return fmt.Sprintf("%s-%s-%s", EksPrefix, clusterName, region)
}
