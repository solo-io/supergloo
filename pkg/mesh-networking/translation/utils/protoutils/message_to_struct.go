package protoutils

import (
	"bytes"

	"github.com/envoyproxy/go-control-plane/pkg/conversion"
	gogojsonpb "github.com/gogo/protobuf/jsonpb"
	gogoproto "github.com/gogo/protobuf/proto"
	gogotypes "github.com/gogo/protobuf/types"
	golangjsonpb "github.com/golang/protobuf/jsonpb"
	golangproto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/rotisserie/eris"
)

// marshal a message to golang/protobuf struct
func GogoMessageToGolangStruct(msg gogoproto.Message) (*pstruct.Struct, error) {
	return conversion.MessageToStruct(msg)
}

// marshal a message to a gogo/protobuf struct
func GogoMessageToGogoStruct(msg gogoproto.Message) (*gogotypes.Struct, error) {
	if msg == nil {
		return nil, eris.New("nil message")
	}

	buf := &bytes.Buffer{}
	if err := (&gogojsonpb.Marshaler{OrigName: true}).Marshal(buf, msg); err != nil {
		return nil, err
	}

	pbs := &gogotypes.Struct{}
	if err := gogojsonpb.Unmarshal(buf, pbs); err != nil {
		return nil, err
	}

	return pbs, nil
}

func GolangMessageToGogoStruct(msg golangproto.Message) (*gogotypes.Struct, error) {
	if msg == nil {
		return nil, eris.New("nil message")
	}
	// Marshal to bytes using golang protobuf
	buf := &bytes.Buffer{}
	if err := (&golangjsonpb.Marshaler{OrigName: true}).Marshal(buf, msg); err != nil {
		return nil, err
	}
	// Unmarshal to gogo protobuf Struct using gogo unmarshaller
	pbs := &gogotypes.Struct{}
	if err := gogojsonpb.Unmarshal(buf, pbs); err != nil {
		return nil, err
	}
	return pbs, nil
}

// MessageToAnyWithError converts from proto message to proto Any
func MessageToAnyWithError(msg golangproto.Message) (*any.Any, error) {
	b := golangproto.NewBuffer(nil)
	b.SetDeterministic(true)
	err := b.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return &any.Any{
		TypeUrl: "type.googleapis.com/" + golangproto.MessageName(msg),
		Value:   b.Bytes(),
	}, nil
}
