package fieldutils

import (
	"fmt"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

type FieldConflictError struct {
	Field     interface{}
	Owner     ezkube.ResourceId
	OwnerType ezkube.Object
	Priority  int32
}

func (e FieldConflictError) Error() string {
	return fmt.Sprintf("field %v is alredy owned by %T %s (priority %v)", e.Field, e.OwnerType, sets.Key(e.Owner), e.Priority)
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
	//
	// field must be a pointer to the field.
	RegisterFieldOwner(obj ezkube.Object, field interface{}, owner ezkube.ResourceId, ownerType ezkube.Object, priority int32) error

	// gets all the owners who share an object.
	GetRegisteredOwners(obj ezkube.Object) []FieldOwner
}

type ownershipRegistry struct {
	objOwners   map[string][]FieldOwner
	fieldOwners map[interface{}]FieldOwner
}

func NewOwnershipRegistry() FieldOwnershipRegistry {
	return &ownershipRegistry{fieldOwners: map[interface{}]FieldOwner{}}
}

type FieldOwner struct {
	owner     ezkube.ResourceId
	ownerType ezkube.Object
	priority  int32
}

func (o *ownershipRegistry) RegisterFieldOwner(obj ezkube.Object, field interface{}, owner ezkube.ResourceId, ownerType ezkube.Object, priority int32) error {
	previousOwner, exists := o.fieldOwners[field]
	if exists && previousOwner.priority >= priority {
		return FieldConflictError{
			Owner:     previousOwner.owner,
			OwnerType: previousOwner.ownerType,
			Priority:  previousOwner.priority,
			Field:     field,
		}
	}
	newOwner := FieldOwner{
		owner:     owner,
		ownerType: ownerType,
		priority:  priority,
	}
	o.fieldOwners[field] = newOwner
	o.registerObjOwner(obj, newOwner)
	return nil
}

func (o *ownershipRegistry) registerObjOwner(obj ezkube.Object, newOwner FieldOwner) {
	key := typedObjectKey(obj)
	for _, owner := range o.objOwners[key] {
		if typedOwnerKey(owner) == typedOwnerKey(newOwner) {
			// prevent duplicates
			return
		}
	}
	o.objOwners[key] = append(o.objOwners[key], newOwner)
}

func (o *ownershipRegistry) GetRegisteredOwners(obj ezkube.Object) []FieldOwner {
	return o.objOwners[typedObjectKey(obj)]
}

func typedObjectKey(obj ezkube.Object) string {
	return fmt.Sprintf("%v.%v.%v.%T", obj.GetName(), obj.GetNamespace(), obj.GetClusterName(), obj)
}

func typedOwnerKey(owner FieldOwner) string {
	return fmt.Sprintf("%v.%v.%T", owner.owner.GetName(), owner.owner.GetNamespace(), owner.ownerType)
}
