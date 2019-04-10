package utils

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func ValidateMeshGroups(meshes v1.MeshList, meshGroups v1.MeshGroupList, resourceErrs reporter.ResourceErrors) {
	for _, mg := range meshGroups {
		for _, ref := range mg.Meshes {
			if ref == nil {
				resourceErrs.AddError(mg, errors.Errorf("ref cannot be nil"))
				continue
			}
			if _, err := meshes.Find(ref.Strings()); err != nil {
				resourceErrs.AddError(mg, err)
				continue
			}
		}
	}
}
