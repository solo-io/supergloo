package equalityutils

import (
	"reflect"
)

// TODO(ilackarms): consider optimizations for this function,
// currently it's used to determine whether two Policies
// produce equivalent configs (to decide whether there might be a conflict).
func Equals(v1, v2 interface{}) bool {
	// TODO(EItanya): Should this be `proto.Equal` since these values are usually proto messages
	return reflect.DeepEqual(v1, v2)
}
