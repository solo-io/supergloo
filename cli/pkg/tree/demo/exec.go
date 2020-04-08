package demo

import (
	"io"
	"os/exec"
)

func NewCommandLineRunner(in io.Reader, out io.Writer) CommandLineRunner {
	return &runner{
		in:  in,
		out: out,
	}
}

type runner struct {
	in  io.Reader
	out io.Writer
}

func (r *runner) Run(cmd string, args ...string) error {
	execCmd := exec.Command(cmd, args...)
	execCmd.Stdin = r.in
	execCmd.Stdout = r.out
	execCmd.Stderr = r.out
	return execCmd.Run()
}
