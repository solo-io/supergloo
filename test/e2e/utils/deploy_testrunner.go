package utils

func DeployTestRunner(namespace string) error {
	return KubectlApply(namespace, TestRunnerYaml)
}

func DeployTestRunnerWithIstioInject(namespace, istioNamespace string) error {
	injected, err := IstioInject(istioNamespace, TestRunnerYaml)
	if err != nil {
		return err
	}

	return KubectlApply(namespace, injected)
}

func DeployTestRunnerWithLinkerdInject(namespace string) error {
	injected, err := LinkerdInject(TestRunnerYaml)
	if err != nil {
		return err
	}

	return KubectlApply(namespace, injected)
}

const TestRunnerYaml = `
##################################################################################################
# Testrunner
##################################################################################################

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

---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: testrunner
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: testrunner
        supergloo: testrunner
    spec:
      serviceAccountName: testrunner
      containers:
        - name: testrunner
          image: soloio/testrunner:testing-8671e8b9
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
          command:
            - sleep
            - "36000"
      restartPolicy: Always
`
