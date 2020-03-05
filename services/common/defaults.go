package common

import (
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
)

var (
	OwnerLabels = map[string]string{
		"owner": "service-mesh-hub",
	}

	/*
		A name used to represent the local cluster when setting up multi cluster watches.
		The empty string "" represents the local cluster, however, a controller cannot have
		an empty string as a name, so this is meant to be an known alternative which will not
		clash with any kubernetes names.
	*/
	LocalClusterName = "local"
)

// Return true if userSuppliedRef's cluster name is the empty string and computedRef's cluster name is LocalClusterName
// This asymmetry is a result of allowing the user to omit the cluster name as shorthand for the cluster on which ServiceMeshHub
// is installed, but our computed configuration uses LocalClusterName as the cluster name.
func AreResourcesOnLocalCluster(userSuppliedRef *core_types.ResourceRef, computedRef *core_types.ResourceRef) bool {
	return userSuppliedRef.GetCluster().GetValue() == "" && computedRef.GetCluster().GetValue() == LocalClusterName
}
