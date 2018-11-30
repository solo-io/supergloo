package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

// TODO: ensure this is the right separator (doesn't seem to be documented)
const portListSeparator = ","

// todo:
// run everything in the test with kube injection on - add injection to all the Deploy commands
// or turn on auto inject for the ns
// then simply run cURL from the testrunner against reviews... try a curl w/ guaranteed response
//
func IstioInject(istioNamespace, input string) (string, error) {
	cmd := exec.Command("istioctl", "kube-inject", "-i", istioNamespace, "-f", "-")
	cmd.Stdin = bytes.NewBuffer([]byte(input))
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, "kube inject failed: %v", output.String())
	}
	return output.String(), nil
}

func AwsAppMeshInject(pod *v1.Pod, meshName, virtualNodeName, awsRegion string, applicationPorts []uint32) {
	pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{
		Name:  "envoy",
		Image: "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta",
		SecurityContext: &v1.SecurityContext{
			RunAsUser: aws.Int64(1337),
		},
		Env: []v1.EnvVar{
			{
				Name:  "APPMESH_VIRTUAL_NODE_NAME",
				Value: "mesh/" + meshName + "/virtualNode/" + virtualNodeName,
			},
			{
				Name:  "ENVOY_LOG_LEVEL",
				Value: "info",
			},
			{
				Name:  "AWS_REGION",
				Value: awsRegion,
			},
		},
	})
	var stringPorts []string
	for _, port := range applicationPorts {
		stringPorts = append(stringPorts, fmt.Sprintf("%v", port))
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, v1.Container{
		Name:  "proxyinit",
		Image: "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest",
		SecurityContext: &v1.SecurityContext{
			Capabilities: &v1.Capabilities{
				Add: []v1.Capability{
					"NET_ADMIN",
				},
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "APPMESH_START_ENABLED",
				Value: "1",
			},
			{
				Name:  "APPMESH_IGNORE_UID",
				Value: "1337",
			},
			{
				Name:  "APPMESH_ENVOY_INGRESS_PORT",
				Value: "15000",
			},
			{
				Name:  "APPMESH_ENVOY_EGRESS_PORT",
				Value: "15001",
			},
			{
				Name:  "APPMESH_APP_PORTS",
				Value: strings.Join(stringPorts, portListSeparator),
			},
			{
				Name:  "APPMESH_EGRESS_IGNORED_IP",
				Value: "169.254.169.254",
			},
			{
				Name:  "APPMESH_START_ENABLED",
				Value: "mesh/" + meshName + "/virtualNode/" + virtualNodeName,
			},
			{
				Name:  "AWS_REGION",
				Value: awsRegion,
			},
		},
	})
}
