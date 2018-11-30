package utils

import (
	"fmt"
	"strings"

	"k8s.io/client-go/rest"
)

func DeployTestRunner(cfg *rest.Config, namespace string) error {
	return DeployFromYamlWithIstioInject(cfg, namespace, TestRunnerYaml)
}

func DeployTestRunnerAppMesh(cfg *rest.Config, namespace, meshName, virtualNodeName, awsRegion string, applicationPorts []uint32) error {
	var stringPorts []string
	for _, port := range applicationPorts {
		stringPorts = append(stringPorts, fmt.Sprintf("%v", port))
	}
	injectedYaml := fmt.Sprintf(TestRunnerAwsAppMeshYaml, meshName, virtualNodeName, awsRegion, strings.Join(stringPorts, ","))
	return DeployFromYamlWithIstioInject(cfg, namespace, injectedYaml)
}

const TestRunnerYaml = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    gloo: testrunner
  name: testrunner
spec:
  containers:
  - image: soloio/testrunner:testing-8671e8b9
    imagePullPolicy: IfNotPresent
    command:
      - sleep
      - "36000"
    name: testrunner
  restartPolicy: Always`

const TestRunnerAwsAppMeshYaml = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    gloo: testrunner
  name: testrunner
spec:
  restartPolicy: Always
  containers:
  - name: testrunner
    image: soloio/testrunner:testing-8671e8b9
    imagePullPolicy: IfNotPresent
    command:
      - sleep
      - "36000" 
  - name: envoy
        image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
        securityContext:
          runAsUser: 1337
        env:
          - name: "APPMESH_VIRTUAL_NODE_NAME"
            value: "mesh/%v/virtualNode/%v"
          - name: "ENVOY_LOG_LEVEL"
            value: "info"
          - name: "AWS_REGION"
            value: "%v"
  initContainers:
    - name: proxyinit
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      securityContext:
        capabilities:
          add: 
            - NET_ADMIN
      env:
      - name: "APPMESH_START_ENABLED"
        value: "1"
      - name: "APPMESH_IGNORE_UID"
        value: "1337"
      - name: "APPMESH_ENVOY_INGRESS_PORT"
        value: "15000"
      - name: "APPMESH_ENVOY_EGRESS_PORT"
        value: "15001"
      - name: "APPMESH_APP_PORTS"
        value: "%v"
      - name: "APPMESH_EGRESS_IGNORED_IP"
        value: "169.254.169.254"
`
