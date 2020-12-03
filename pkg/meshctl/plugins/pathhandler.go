package plugins

import (
	"fmt"
	"os"
	"os/exec"
)

// PathHandler attempts to load a plugin from the system path.
type PathHandler struct {
	prefix string
}

func NewPathHandler(prefix string) *PathHandler {
	return &PathHandler{prefix}
}

func (ph PathHandler) Lookup(pluginName string) (Plugin, bool) {
	binaryName := fmt.Sprintf("%s-%s", ph.prefix, pluginName)
	binaryPath, err := exec.LookPath(binaryName)
	if err == exec.ErrNotFound {
		return nil, false
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error searching system path: %s\n", err.Error())
		return nil, false
	}

	return NewBinaryPlugin(binaryPath), true
}
