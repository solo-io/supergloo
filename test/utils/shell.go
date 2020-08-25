package utils

import (
	"bytes"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
)

func RunShell(c string) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "[%v] command FAILED: %v", c, err)
		return
	}
	fmt.Fprintf(GinkgoWriter, "[%v] command result: \n%v", c, buf.String())
}
