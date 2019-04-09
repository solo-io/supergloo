package utils

import (
	"strings"

	"github.com/solo-io/go-utils/kubeutils"

	"github.com/solo-io/go-utils/stringutils"
)

func SubsetName(labels map[string]string) string {
	keys, values := stringutils.KeysAndValues(labels)
	name := ""
	for i := range keys {
		name += keys[i] + "-" + values[i] + "-"
	}
	name = strings.TrimSuffix(name, "-")
	return kubeutils.SanitizeName(name)
}
