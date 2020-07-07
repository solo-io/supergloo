package protoutils

import "github.com/gogo/protobuf/proto"

// TODO(ilackarms): replace this equality function with something with better performance,
// consider using code generation on external proto types.
func Equals(m1, m2 proto.Message) bool {
	return proto.Equal(m1, m2)
}