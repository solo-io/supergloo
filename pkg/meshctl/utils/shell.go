package utils

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

func RunShell(c string, writer io.Writer) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(writer, "[%v] command FAILED: %v", c, err)
		return
	}
	fmt.Fprintf(writer, "[%v] command result: \n%v", c, buf.String())
}
