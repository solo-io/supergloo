package utils

import (
	"os"
	"path/filepath"

	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/codegen/render"
	"github.com/solo-io/skv2/codegen/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	perm = 0644
)

type Manifest struct {
	filename string
}

func NewManifest(filename string) (Manifest, error) {
	manifest := Manifest{filename: filepath.Join(filepath.Dir(util.GoModPath()), "test/e2e/"+filename)}
	err := manifest.CreateOrTruncate()
	if err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (m Manifest) KubeApply(namespace string) error {
	return testutils.Kubectl("apply", "-n="+namespace, "-f="+m.filename)
}

func (m Manifest) KubeDelete(namespace string) error {
	return testutils.Kubectl("delete", "-n="+namespace, "-f="+m.filename)
}

// Same as KubeDelete but ignore errors in case the test has already cleaned up associated resources.
// This method is just a safeguard for ensuring a clean test slate between unit tests.
func (m Manifest) Cleanup(namespace string) {
	testutils.Kubectl("delete", "-n="+namespace, "-f="+m.filename)
}

func (m Manifest) AppendResources(resources ...metav1.Object) error {
	// use skv2 libraries to write the resources as yaml
	manifest, err := render.ManifestsRenderer{
		AppName: "bookinfo-policies",
		ResourceFuncs: map[render.OutFile]render.MakeResourceFunc{
			render.OutFile{}: func(group render.Group) ([]metav1.Object, error) {
				return resources, nil
			},
		},
	}.RenderManifests(model.Group{RenderManifests: true})
	if err != nil {
		return err
	}

	f, err := os.OpenFile(m.filename, os.O_APPEND|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(manifest[0].Content + "\n---\n"))
	if err != nil {
		return err
	}

	return nil
}

func (m Manifest) CreateOrTruncate() error {
	_, err := os.OpenFile(m.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	return err
}
