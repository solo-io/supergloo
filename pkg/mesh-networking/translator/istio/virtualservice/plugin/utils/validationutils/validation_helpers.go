package validationutils

import (
	"github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
)

func ValidateDuration(duration *types.Duration) error {
	if duration.GetSeconds() < 0 || (duration.GetSeconds() == 0 && duration.GetNanos() < 1000000) {
		return eris.New("Duration must be >= 1 millisecond")
	}
	return nil
}
