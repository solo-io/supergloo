package utils

func DeployTestRunner(namespace string) error {
	return KubectlApply(namespace, TestRunnerYaml)
}

func DeployTestRunnerWithInject(namespace, istioNamespace string) error {
	injected, err := IstioInject(istioNamespace, TestRunnerYaml)
	if err != nil {
		return err
	}

	return KubectlApply(namespace, injected)
}

const TestRunnerYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: testrunner
---
apiVersion: v1
kind: Service
metadata:
  name: testrunner
  labels:
    supergloo: testrunner
spec:
  ports:
  - port: 8080
    name: http
  selector:
    supergloo: testrunner
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    supergloo: testrunner
  name: testrunner
spec:
  serviceAccountName: testrunner
  containers:
  - image: soloio/testrunner:testing-8671e8b9
    imagePullPolicy: IfNotPresent
    command:
      - sleep
      - "36000"
    name: testrunner
  restartPolicy: Always`
