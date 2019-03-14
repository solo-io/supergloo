package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/test/setup"
)

func TestRunnerCurlEventuallyShouldRespond(ctx context.Context, testrunnerNamespace string, opts setup.CurlOpts, substr string, timeout time.Duration) {
	// for some useful-ish output
	tick := time.Tick(timeout / 10)
	gomega.EventuallyWithOffset(1, func() string {
		res, err := TestRunnerCurl(testrunnerNamespace, opts)
		if err != nil {
			res = err.Error()
		}
		select {
		default:
			break
		case <-tick:
			contextutils.LoggerFrom(ctx).Infof("running: %v\nwant %v\nhave: %s", opts, substr, res)
		case <-ctx.Done():
			return ""
		}
		if strings.Contains(res, substr) {
			contextutils.LoggerFrom(ctx).Infof("success: %v", res)
		}
		return res
	}, timeout, "2s").Should(gomega.ContainSubstring(substr))
}

func CurlArgs(opts setup.CurlOpts) []string {
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
	return args
}

func TestRunnerCurl(namespace string, opts setup.CurlOpts) (string, error) {
	args := CurlArgs(opts)
	return setup.TestRunner(namespace, args...)
}
