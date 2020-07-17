package protoutils

import (
	"bytes"

	"github.com/envoyproxy/go-control-plane/pkg/conversion"
	pstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
)

// marshal a message to golang/protobuf struct
func MessageToGolangStruct(msg proto.Message) (*pstruct.Struct, error) {
	return conversion.MessageToStruct(msg)
}

// marshal a message to a gogo/protobuf struct
func MessageToGogoStruct(msg proto.Message) (*types.Struct, error) {
	if msg == nil {
		return nil, eris.New("nil message")
	}

	buf := &bytes.Buffer{}
	if err := (&jsonpb.Marshaler{OrigName: true}).Marshal(buf, msg); err != nil {
		return nil, err
	}

	pbs := &types.Struct{}
	if err := jsonpb.Unmarshal(buf, pbs); err != nil {
		return nil, err
	}

	return pbs, nil
}
