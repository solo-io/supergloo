package aws_utils

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/rotisserie/eris"
	k8s_core_types "k8s.io/api/core/v1"
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
)

type AppMeshPod struct {
	AwsAccountID    string
	Region          string
	AppMeshName     string
	VirtualNodeName string
}

// iterate through pod's containers and check for one with name containing "appmesh" and "proxy"
// if true, return inferred AppMesh name
func ScanPodForAppMesh(pod *k8s_core_types.Pod) (*AppMeshPod, error) {
	var err error
	var awsAccountID, region, appMeshName, virtualNodeName string
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Image, "appmesh") && strings.Contains(container.Image, "envoy") {
			for _, env := range container.Env {
				if env.Name == AppMeshVirtualNodeEnvVarName {
					// Value takes format "mesh/<appmesh-mesh-name>/virtualNode/<virtual-node-name>"
					// VirtualNodeName value has significance for AppMesh functionality, reference:
					// https://docs.aws.amazon.com/eks/latest/userguide/mesh-k8s-integration.html
					appMeshName = strings.Split(env.Value, "/")[1]
					virtualNodeName = strings.Split(env.Value, "/")[3]
				} else if env.Name == AppMeshRegionEnvVarName {
					region = env.Value
				} else if env.Name == AppMeshRoleArnEnvVarName {
					awsAccountID, err = ParseAwsAccountID(env.Value)
					if err != nil {
						return nil, err
					}
				}
			}
			return &AppMeshPod{
				AwsAccountID:    awsAccountID,
				Region:          region,
				AppMeshName:     appMeshName,
				VirtualNodeName: virtualNodeName,
			}, nil
		}
	}
	return nil, nil
}

func ParseAwsAccountID(arnString string) (string, error) {
	parse, err := arn.Parse(arnString)
	if err != nil {
		return "", ARNParseError(err, arnString)
	}
	return parse.AccountID, nil
}
