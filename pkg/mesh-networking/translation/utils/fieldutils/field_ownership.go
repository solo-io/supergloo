package fieldutils

import (
	"fmt"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

type FieldConflictError struct {
	Field    interface{}
	Owner    ezkube.ResourceId
	Priority int32
}

func (e FieldConflictError) Error() string {
	return fmt.Sprintf("field %v is alredy owned by %T %s (priority %v)", e.Field, e.Owner, sets.Key(e.Owner), e.Priority)
}

// an FieldOwnershipRegistry tracks the ownership of individual object fields.
// this is used to track e.g. which TrafficPolicy
// Note that the current implementation of the ownership registry
// is intentionally non-threadsafe. an FieldOwnershipRegistry should
// only be called within a single threaded context (i.e. in a translation loop)
type FieldOwnershipRegistry interface {
	// Registers an owner with a given priority for the given field.
	//
	// If an owner with a higher or equal priority exists for the field,
	// a ConflictError containing the previous owner and its priority are returned.
	RegisterFieldOwner(field interface{}, owner ezkube.ResourceId, priority int32) error
}

type ownershipRegistry struct {
	fieldOwners map[interface{}]fieldOwner
}

func NewOwnershipRegistry() FieldOwnershipRegistry {
	return &ownershipRegistry{fieldOwners: map[interface{}]fieldOwner{}}
}

type fieldOwner struct {
	owner    ezkube.ResourceId
	priority int32
}

func (o ownershipRegistry) RegisterFieldOwner(field interface{}, owner ezkube.ResourceId, priority int32) error {
	previousOwner, exists := o.fieldOwners[field]
	if exists && previousOwner.priority >= priority {
		return FieldConflictError{
			Owner:    previousOwner.owner,
			Priority: previousOwner.priority,
			Field:    field,
		}
	}
	o.fieldOwners[field] = fieldOwner{
		owner:    owner,
		priority: priority,
	}
	return nil
}
