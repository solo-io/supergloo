package v1alpha1

// Status of the resource
type Status string

const (
	// StatusPending is set when the resource is pending creation
	StatusPending Status = "Pending"
	// StatusCreated is set when the resource has been successfully created
	StatusCreated Status = "Created"
	// StatusError is set when a fatal error has occured during the creation of
	// the resource
	StatusError Status = "Error"
)
