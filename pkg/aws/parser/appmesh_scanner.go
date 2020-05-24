package aws_utils

import (
	"strings"

	"github.com/rotisserie/eris"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Used to infer parent AppMesh Mesh name
	AppMeshVirtualNodeEnvVarName = "APPMESH_VIRTUAL_NODE_NAME"
	AppMeshRegionEnvVarName      = "AWS_REGION"
	AppMeshRoleArnEnvVarName     = "AWS_ROLE_ARN"
)

var (
	ARNParseError = func(err error, arn string) error {
		return eris.Wrapf(err, "Error parsing ARN: %s", arn)
	}
	EmptyEnvVarValueError = func(envVarName string, podMeta k8s_meta_types.ObjectMeta) error {
		return eris.Errorf("Missing value for env var %s on pod %s, %s", envVarName, podMeta.GetName(), podMeta.GetNamespace())
	}
	UnexpectedVirtualNodeValue = func(virtualNodeEnvVarValue string) error {
		return eris.Errorf("Unexpected value for env var %s: %s", AppMeshVirtualNodeEnvVarName, virtualNodeEnvVarValue)
	}
)

type appMeshScanner struct {
	arnParser ArnParser
}

func NewAppMeshScanner(
	arnParser ArnParser,
) AppMeshScanner {
	return &appMeshScanner{
		arnParser: arnParser,
	}
}

// iterate through pod's containers and check for one with name containing "appmesh" and "proxy"
// if true, return inferred AppMesh name
func (a *appMeshScanner) ScanPodForAppMesh(
	pod *k8s_core_types.Pod,
	awsAccountId AwsAccountId,
) (*AppMeshPod, error) {
	var err error
	var awsAccountIDString, region, appMeshName, virtualNodeName string
	awsAccountIDString = string(awsAccountId)
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Image, "appmesh") && strings.Contains(container.Image, "envoy") {
			for _, env := range container.Env {
				if env.Name == AppMeshVirtualNodeEnvVarName {
					// Value takes format "mesh/<appmesh-mesh-name>/virtualNode/<virtual-node-name>"
					// VirtualNodeName value has significance for AppMesh functionality, reference:
					// https://docs.aws.amazon.com/eks/latest/userguide/mesh-k8s-integration.html
					split := strings.Split(env.Value, "/")
					if len(split) != 4 {
						return nil, UnexpectedVirtualNodeValue(env.Value)
					}
					appMeshName = split[1]
					virtualNodeName = split[3]
				} else if env.Name == AppMeshRegionEnvVarName {
					if env.Value == "" {
						return nil, EmptyEnvVarValueError(env.Name, pod.ObjectMeta)
					}
					region = env.Value
				} else if env.Name == AppMeshRoleArnEnvVarName && awsAccountIDString == "" {
					// Fallback to using env var if AWS account ID wasn't found from aws-auth ConfigMap
					if env.Value == "" {
						return nil, EmptyEnvVarValueError(env.Name, pod.ObjectMeta)
					}
					awsAccountIDString, err = a.arnParser.ParseAccountID(env.Value)
					if err != nil {
						return nil, err
					}
				}
			}
			// If any of the below variables is empty, return error
			return &AppMeshPod{
				AwsAccountID:    awsAccountIDString,
				Region:          region,
				AppMeshName:     appMeshName,
				VirtualNodeName: virtualNodeName,
			}, nil
		}
	}
	return nil, nil
}
