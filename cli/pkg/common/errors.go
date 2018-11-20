package common

const (
	UnknownResourceTypeMsg = "the server doesn't have a resource type \"%v\""
	UnknownOutputFormat    = "unknown output format \"%v\", allowed formats are: \"%v\""

	InvalidOptionFormat      = "invalid format for option: %v. \n Use \"supergloo %s -h\" for more information about available options"
	InvalidMatcherHttpMethod = "invalid HTTP method: %v"

	KubeConfigError = "Error retrieving Kubernetes configuration: %v \n"
)
