package utils

import (
	"os/exec"

	"github.com/solo-io/go-utils/errors"
)

func HelmTemplate(args ...string) (string, error) {
	out, err := exec.Command("helm", append([]string{"template"}, args...)...).CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "helm template failed: %v", string(out))
	}
	return string(out), nil
}
