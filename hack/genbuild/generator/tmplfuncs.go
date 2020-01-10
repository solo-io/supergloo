package generator

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

var GenBuildFuncs = template.FuncMap{
	"uppercase": strings.ToUpper,
	"jsoner":    Jsoner,
	"makefileArg": func(in string, suffix string) string {
		base := strcase.ToScreamingDelimited(in, '_', 0, true)
		if suffix != "" {
			return strings.Join([]string{base, suffix}, "_")
		}
		return base
	},
	"wrap": func(content string) string {
		return fmt.Sprintf("{{%v}}", content)
	},
	"kebab": strcase.ToKebab,
	"docker_gcloud": func() string {
		return gcloudDocker
	},
}

func Jsoner(in interface{}) (string, error) {
	b, err := json.Marshal(in)
	return string(b), err
}
