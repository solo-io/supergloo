package utils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/codegen/render"
	"github.com/solo-io/skv2/codegen/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WriteTestManifest(manifestFile string, objs []metav1.Object) error {
	return writeResourcesToManifest(objs, manifestFile)
}

func writeResourcesToManifest(resources []metav1.Object, filename string) error {
	// use skv2 libraries to write the resources as yaml
	manifest, err := render.ManifestsRenderer{
		AppName: "bookinfo-policies",
		ResourceFuncs: map[render.OutFile]render.MakeResourceFunc{
			render.OutFile{}: func(group render.Group) []metav1.Object {
				return resources
			},
		},
	}.RenderManifests(model.Group{RenderManifests: true})
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(filepath.Dir(util.GoModPath()), filename), []byte(manifest[0].Content), 0644); err != nil {
		return err
	}

	return nil
}
