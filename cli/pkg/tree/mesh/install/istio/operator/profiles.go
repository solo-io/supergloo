package operator

import "k8s.io/apimachinery/pkg/util/sets"

var ValidProfiles = sets.NewString(
	"demo",
	"default",
	"minimal",
	"sds",
	"remote",
)
