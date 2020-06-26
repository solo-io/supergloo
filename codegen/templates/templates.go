package templates

import (
	"io/ioutil"
	"log"
)

var (
	OutputSnapshotTemplateContents = mustRead("codegen/templates/output_snapshot.gotmpl")
	ReconcilerTemplateContents     = mustRead("codegen/templates/reconciler.gotmpl")
)

func mustRead(file string) string {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read file %v", err)
	}
	return string(b)
}
