package setup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

const (
	testrunner = "testrunner"
)

func SetupKubeForTest(namespace string) error {
	context := os.Getenv("KUBECTL_CONTEXT")
	if context == "" {
		current, err := KubectlOut("config", "current-context")
		if err != nil {
			return errors.Wrap(err, "getting currrent context")
		}
		context = strings.TrimSuffix(current, "\n")
	}
	// TODO(yuval-k): this changes the context for the user? can we do this less intrusive? maybe add it to
	// each kubectl command?
	if err := Kubectl("config", "set-context", context, "--namespace="+namespace); err != nil {
		return errors.Wrap(err, "setting context")
	}
	return Kubectl("create", "namespace", namespace)
}

func TeardownKube(namespace string) error {
	return Kubectl("delete", "namespace", namespace)
}
func Kubectl(args ...string) error {
	cmd := exec.Command("kubectl", args...)
	log.Debugf("k command: %v", cmd.Args)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	return cmd.Run()
}

func KubectlOut(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s (%v)", out, err)
	}
	return string(out), err
}

func KubectlOutAsync(args ...string) (*bytes.Buffer, chan struct{}, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Start()
	if err != nil {
		err = fmt.Errorf("%s (%v)", buf.Bytes(), err)
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			cmd.Process.Kill()
		}
	}()
	return buf, done, err
}

// WaitPodsRunning waits for all pods to be running
func WaitPodsRunning(podNames ...string) error {
	finished := func(output string) bool {
		return strings.Contains(output, "Running")
	}
	for _, pod := range podNames {
		if err := WaitPodStatus(pod, "Running", finished); err != nil {
			return err
		}
	}
	return nil
}

// waitPodsTerminated waits for all pods to be terminated
func WaitPodsTerminated(podNames ...string) error {
	for _, pod := range podNames {
		finished := func(output string) bool {
			return !strings.Contains(output, pod)
		}
		if err := WaitPodStatus(pod, "terminated", finished); err != nil {
			return err
		}
	}
	return nil
}

// TestRunner executes a command inside the TestRunner container
func TestRunner(args ...string) (string, error) {
	args = append([]string{"exec", "-i", testrunner, "--"}, args...)
	return KubectlOut(args...)
}

// TestRunnerAsync executes a command inside the TestRunner container
// returning a buffer that can be read from as it executes
func TestRunnerAsync(args ...string) (*bytes.Buffer, chan struct{}, error) {
	args = append([]string{"exec", "-i", testrunner, "--"}, args...)
	return KubectlOutAsync(args...)
}

func WaitPodStatus(pod, status string, finished func(output string) bool) error {
	timeout := time.Second * 20
	interval := time.Millisecond * 1000
	tick := time.Tick(interval)

	log.Debugf("waiting %v for pod %v to be %v...", timeout, pod, status)
	for {
		select {
		case <-time.After(timeout):
			return fmt.Errorf("timed out waiting for %v to be %v", pod, status)
		case <-tick:
			out, err := KubectlOut("get", "pod", "-l", "gloo="+pod)
			if err != nil {
				return fmt.Errorf("failed getting pod: %v", err)
			}
			if strings.Contains(out, "CrashLoopBackOff") {
				out = KubeLogs(pod)
				return errors.Errorf("%v in crash loop with logs %v", pod, out)
			}
			if strings.Contains(out, "ErrImagePull") || strings.Contains(out, "ImagePullBackOff") {
				out, _ = KubectlOut("describe", "pod", "-l", "gloo="+pod)
				return errors.Errorf("%v in ErrImagePull with description %v", pod, out)
			}
			if finished(out) {
				return nil
			}
		}
	}
}

func KubeLogs(pod string) string {
	out, err := KubectlOut("logs", "-l", "gloo="+pod)
	if err != nil {
		out = err.Error()
	}
	return out
}

func WaitNamespaceStatus(namespace, status string, finished func(output string) bool) error {
	timeout := time.Second * 20
	interval := time.Millisecond * 1000
	tick := time.Tick(interval)

	log.Debugf("waiting %v for namespace %v to be %v...", timeout, namespace, status)
	for {
		select {
		case <-time.After(timeout):
			return fmt.Errorf("timed out waiting for %v to be %v", namespace, status)
		case <-tick:
			out, err := KubectlOut("get", "namespace", namespace)
			if err != nil {
				return fmt.Errorf("failed getting pod: %v", err)
			}
			if finished(out) {
				return nil
			}
		}
	}
}

type CurlOpts struct {
	Protocol      string
	Path          string
	Method        string
	Host          string
	Service       string
	CaFile        string
	Body          string
	Headers       map[string]string
	Port          int
	ReturnHeaders bool
}

func CurlEventuallyShouldRespond(opts CurlOpts, substr string, timeout ...time.Duration) {
	t := time.Second * 20
	if len(timeout) > 0 {
		t = timeout[0]
	}
	// for some useful-ish output
	tick := time.Tick(t / 8)
	gomega.Eventually(func() string {
		res, err := Curl(opts)
		if err != nil {
			res = err.Error()
		}
		select {
		default:
			break
		case <-tick:
			log.GreyPrintf("running: %v\nwant %v\nhave: %s", opts, substr, res)
		}
		if strings.Contains(res, substr) {
			log.GreyPrintf("success: %v", res)
		}
		return res
	}, t, "5s").Should(gomega.ContainSubstring(substr))
}

func Curl(opts CurlOpts) (string, error) {
	args := []string{"curl", "-v", "--connect-timeout", "10", "--max-time", "10"}

	if opts.ReturnHeaders {
		args = append(args, "-I")
	}

	if opts.Method != "GET" && opts.Method != "" {
		args = append(args, "-X"+opts.Method)
	}
	if opts.Host != "" {
		args = append(args, "-H", "Host: "+opts.Host)
	}
	if opts.CaFile != "" {
		args = append(args, "--cacert", opts.CaFile)
	}
	if opts.Body != "" {
		args = append(args, "-H", "Content-Type: application/json")
		args = append(args, "-d", opts.Body)
	}
	for h, v := range opts.Headers {
		args = append(args, "-H", fmt.Sprintf("%v: %v", h, v))
	}
	port := opts.Port
	if port == 0 {
		port = 8080
	}
	protocol := opts.Protocol
	if protocol == "" {
		protocol = "http"
	}
	service := opts.Service
	if service == "" {
		service = "test-ingress"
	}
	args = append(args, fmt.Sprintf("%v://%s:%v%s", protocol, service, port, opts.Path))
	log.Debugf("running: curl %v", strings.Join(args, " "))
	return TestRunner(args...)
}
