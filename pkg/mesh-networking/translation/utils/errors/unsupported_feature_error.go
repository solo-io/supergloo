package errors

import (
	"fmt"

	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

type UnsupportedFeatureError struct {
	resource  ezkube.ResourceId
	fieldName string
	reason    string
}

func NewUnsupportedFeatureError(resource ezkube.ResourceId, fieldName, reason string) error {
	return &UnsupportedFeatureError{
		resource:  resource,
		fieldName: fieldName,
		reason:    reason,
	}
}

func (u *UnsupportedFeatureError) Error() string {
	return fmt.Sprintf(
		"Unsupported feature %s used on resource %T <%s>. %s",
		u.fieldName,
		u.resource,
		sets.Key(u.resource),
		u.reason,
	)
}
