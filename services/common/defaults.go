package common

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
	LocalClusterName = "$local$"
)
