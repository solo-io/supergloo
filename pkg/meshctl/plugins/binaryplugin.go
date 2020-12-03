package plugins

import (
	"io"
	"os"
	"os/exec"
)

// BinaryPlugin is run via an executable binary.
type BinaryPlugin struct {
	// Where to read input
	In io.Reader

	// Where to write output
	Out io.Writer

	// Write to write error information
	Err io.Writer

	path string
}

func NewBinaryPlugin(path string) *BinaryPlugin {
	return &BinaryPlugin{In: os.Stdin, Out: os.Stdout, Err: os.Stderr, path: path}
}

func (bp BinaryPlugin) Run(args []string) error {
	// cmd := exec.Command(bp.path, append([]string{bp.path}, args...)...)
	cmd := exec.Command(bp.path, args...)
	cmd.Stdin = bp.In
	cmd.Stdout = bp.Out
	cmd.Stderr = bp.Err
	return cmd.Run()
}
