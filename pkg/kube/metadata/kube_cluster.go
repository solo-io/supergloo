package metadata

import "fmt"

const (
	EksPrefix = "eks"
)

func BuildEksClusterName(clusterName, region string) string {
	return fmt.Sprintf("%s-%s-%s", EksPrefix, clusterName, region)
}
