package version

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/solo-io/mesh-projects/pkg/version"
)

func ReportVersion(out io.Writer) error {
	versionInfo := map[string]string{
		"version": version.Version,
	}

	bytes, err := json.Marshal(versionInfo)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "%s\n", string(bytes))
	return nil
}
