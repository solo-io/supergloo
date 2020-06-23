package output

import (
	"github.com/rotisserie/eris"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshots/output"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"sort"
)

// the Applier applies the output snapshot to the backing persistence store (i.e. Kube CRDs)
type Applier interface {
	Apply(output output.Snapshot) error
}

var MissingRequiredLabelError = func(labelKey, resourceKind string, obj ezkube.ResourceId) error {
	return eris.Errorf("expected label %v not on labels of %v %v", labelKey, resourceKind, sets.Key(obj))
}

func partitionMeshesByLabel(labelKey string, set v1alpha1sets.MeshSet) ([]output.LabeledMeshSet, error) {
	setsByLabel := map[string]v1alpha1sets.MeshSet{}

	for _, obj := range set.List() {
		if obj.Labels == nil {
			return nil, MissingRequiredLabelError(labelKey, "Mesh", obj)
		}
		labelValue := obj.Labels[labelKey]
		if labelValue == "" {
			return nil, MissingRequiredLabelError(labelKey, "Mesh", obj)
		}

		setForValue, ok := setsByLabel[labelValue]
		if !ok {
			setForValue = v1alpha1sets.NewMeshSet()
			setsByLabel[labelValue] = setForValue
		}
		setForValue.Insert(obj)
	}

	// partition by label key
	var partitionedMeshes []output.LabeledMeshSet

	for labelValue, setForValue := range setsByLabel {
		labels := map[string]string{labelKey: labelValue}

		partitionedSet, err := output.NewLabeledMeshSet(setForValue, labels)
		if err != nil {
			return nil, err
		}

		partitionedMeshes = append(partitionedMeshes, partitionedSet)
	}

	// sort for idempotency
	sort.SliceStable(partitionedMeshes, func(i, j int) bool {
		leftLabelValue := partitionedMeshes[i].Labels()[labelKey]
		rightLabelValue := partitionedMeshes[j].Labels()[labelKey]
		return leftLabelValue < rightLabelValue
	})

	return partitionedMeshes, nil
}
