// ilackarms: This file contains more than just proto-utils at this point. Should be split, or
// moved to a general serialization util package

package protoutils

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

var jsonpbMarshaler = &jsonpb.Marshaler{OrigName: false}
var jsonpbMarshalerEmitZeroValues = &jsonpb.Marshaler{OrigName: false, EmitDefaults: true}

// this function is designed for converting go object (that is not a proto.Message) into a
// pb Struct, based on json struct tags
func MarshalStruct(m proto.Message) (*types.Struct, error) {
	data, err := MarshalBytes(m)
	if err != nil {
		return nil, err
	}
	var pb types.Struct
	err = jsonpb.UnmarshalString(string(data), &pb)
	return &pb, err
}

func MarshalStructEmitZeroValues(m proto.Message) (*types.Struct, error) {
	data, err := MarshalBytesEmitZeroValues(m)
	if err != nil {
		return nil, err
	}
	var pb types.Struct
	err = jsonpb.UnmarshalString(string(data), &pb)
	return &pb, err
}

func UnmarshalStruct(structuredData *types.Struct, into interface{}) error {
	if structuredData == nil {
		return errors.New("cannot unmarshal nil proto struct")
	}
	strData, err := jsonpbMarshaler.MarshalToString(structuredData)
	if err != nil {
		return err
	}
	data := []byte(strData)
	return json.Unmarshal(data, into)
}

func UnmarshalBytes(data []byte, into proto.Message) error {
	return jsonpb.Unmarshal(bytes.NewBuffer(data), into)
}

func UnmarshalYaml(data []byte, into proto.Message) error {
	jsn, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		return err
	}

	return jsonpb.Unmarshal(bytes.NewBuffer(jsn), into)
}

func MarshalBytes(pb proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := jsonpbMarshaler.Marshal(buf, pb)
	return buf.Bytes(), err
}

func MarshalBytesEmitZeroValues(pb proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := jsonpbMarshalerEmitZeroValues.Marshal(buf, pb)
	return buf.Bytes(), err
}

func MarshalMap(from proto.Message) (map[string]interface{}, error) {
	data, err := MarshalBytes(from)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return m, err
}

func MarshalMapEmitZeroValues(from proto.Message) (map[string]interface{}, error) {
	data, err := MarshalBytesEmitZeroValues(from)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return m, err
}

func UnmarshalMap(m map[string]interface{}, into proto.Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return UnmarshalBytes(data, into)
}

// ilackarms: help come up with a better name for this please
// values in stringMap are yaml encoded or error
// used by configmap resource client
func MapStringStringToMapStringInterface(stringMap map[string]string) (map[string]interface{}, error) {
	interfaceMap := make(map[string]interface{})
	for k, strVal := range stringMap {
		var interfaceVal interface{}
		if err := yaml.Unmarshal([]byte(strVal), &interfaceVal); err != nil {
			return nil, errors.Errorf("%v cannot be parsed as yaml")
		} else {
			interfaceMap[k] = interfaceVal
		}
	}
	return interfaceMap, nil
}

// reverse of previous
func MapStringInterfaceToMapStringString(interfaceMap map[string]interface{}) (map[string]string, error) {
	stringMap := make(map[string]string)
	for k, interfaceVal := range interfaceMap {
		yml, err := yaml.Marshal(interfaceVal)
		if err != nil {
			return nil, errors.Wrapf(err, "map values must be serializable to json")
		}
		stringMap[k] = string(yml)
	}
	return stringMap, nil
}
