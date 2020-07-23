package fieldutils

import (
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/resourceidutils"
	"github.com/solo-io/skv2/pkg/ezkube"
	"k8s.io/apimachinery/pkg/util/sets"
)

type FieldConflictError struct {
	Field     interface{}
	Owners    []ezkube.ResourceId
	OwnerType ezkube.Object
	Priority  int32
}

func (e FieldConflictError) Error() string {
	return fmt.Sprintf("field %v is already owned by %T %s (priority %v)", e.Field, e.OwnerType, resourceidutils.ResourceIdsToString(e.Owners), e.Priority)
}

// an FieldOwnershipRegistry tracks the ownership of individual object fields.
// this is used to track e.g. which TrafficPolicy
// Note that the current implementation of the ownership registry
// is intentionally non-threadsafe. an FieldOwnershipRegistry should
// only be called within a single threaded context (i.e. in a translation loop)
type FieldOwnershipRegistry interface {
	// Registers ownership with a given priority for the given field.
	//
	// If an ownership with a higher or equal priority exists for the field,
	// a ConflictError containing the previous owner and its priority are returned.
	//
	// field must be a pointer to the field.
	RegisterFieldOwnership(obj ezkube.Object, field interface{}, owners []ezkube.ResourceId, ownerType ezkube.Object, priority int32) error

	// gets all the ownerships who share an object.
	GetRegisteredOwnerships(obj ezkube.Object) []FieldOwnership
}

type ownershipRegistry struct {
	objOwners   map[string][]FieldOwnership
	fieldOwners map[interface{}]FieldOwnership
}

func NewOwnershipRegistry() FieldOwnershipRegistry {
	return &ownershipRegistry{
		objOwners: map[string][]FieldOwnership{},
		fieldOwners: map[interface{}]FieldOwnership{},
	}
}

type FieldOwnership struct {
	owners    []ezkube.ResourceId
	ownerType ezkube.Object
	priority  int32
}

func (o *ownershipRegistry) RegisterFieldOwnership(obj ezkube.Object, field interface{}, owners []ezkube.ResourceId, ownerType ezkube.Object, priority int32) error {
	previousOwner, exists := o.fieldOwners[field]
	if exists && previousOwner.priority >= priority {
		return FieldConflictError{
			Owners:    previousOwner.owners,
			OwnerType: previousOwner.ownerType,
			Priority:  previousOwner.priority,
			Field:     field,
		}
	}
	newOwner := FieldOwnership{
		owners:    owners,
		ownerType: ownerType,
		priority:  priority,
	}
	o.fieldOwners[field] = newOwner
	o.registerObjOwner(obj, newOwner)
	return nil
}

func (o *ownershipRegistry) registerObjOwner(obj ezkube.Object, newOwner FieldOwnership) {
	key := typedObjectKey(obj)
	for _, owner := range o.objOwners[key] {
		if typedOwnerSet(owner).Equal(typedOwnerSet(newOwner)) {
			// prevent duplicates
			return
		}
	}
	o.objOwners[key] = append(o.objOwners[key], newOwner)
}

func (o *ownershipRegistry) GetRegisteredOwnerships(obj ezkube.Object) []FieldOwnership {
	return o.objOwners[typedObjectKey(obj)]
}

func typedObjectKey(obj ezkube.Object) string {
	return fmt.Sprintf("%v.%v.%v.%T", obj.GetName(), obj.GetNamespace(), obj.GetClusterName(), obj)
}

func typedOwnerSet(fieldOwner FieldOwnership) sets.String {
	set := sets.NewString()
	for _, owner := range fieldOwner.owners {
		set.Insert(fmt.Sprintf("%v.%v.%T", owner.GetName(), owner.GetNamespace(), fieldOwner.ownerType))
	}
	return set
}
