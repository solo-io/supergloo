package metadata

import (
	"fmt"
	"strings"
)

const (
	AppMeshNamePrefix = "appmesh"
)

func BuildAppMeshName(appMeshName, region, awsAccountId string) string {
	return sanitize(fmt.Sprintf("%s-%s-%s-%s",
		AppMeshNamePrefix,
		appMeshName,
		region,
		awsAccountId,
	))
}

// AppMesh entity names only contain "Alphanumeric characters, dashes, and underscores are allowed." (taken from AppMesh GUI)
// So just replace underscores with a k8s name friendly delimiter
// TODO surface this in the AppMesh portion of user-facing docs
func sanitize(appmeshName string) string {
	return strings.ReplaceAll(appmeshName, "_", "-")
}
