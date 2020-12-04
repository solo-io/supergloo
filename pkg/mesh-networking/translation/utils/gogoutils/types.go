package gogoutils

import (
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/duration"
)

func DurationProtoToGogo(pr *duration.Duration) *types.Duration {
	var ret *types.Duration
	if pr != nil {
		ret = &types.Duration{
			Seconds: pr.GetSeconds(),
			Nanos:   pr.GetNanos(),
		}
	}
	return ret
}

func DurationGogoToProto(pr *types.Duration) *duration.Duration {
	var ret *duration.Duration
	if pr != nil {
		ret = &duration.Duration{
			Seconds: pr.GetSeconds(),
			Nanos:   pr.GetNanos(),
		}
	}
	return ret
}
