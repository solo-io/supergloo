package utils

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/solo-io/supergloo/cli/pkg/cmd"
)

func Supergloo(args string) error {
	app := cmd.SuperglooCli("test")
	app.SetArgs(strings.Split(args, " "))
	return app.Execute()
}

func SuperglooOut(args string) (string, error) {
	stdOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	app := cmd.SuperglooCli("test")
	app.SetArgs(strings.Split(args, " "))
	err = app.Execute()

	outC := make(chan string)

	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = stdOut // restoring the real stdout
	out := <-outC

	return strings.TrimSuffix(out, "\n"), nil
}
