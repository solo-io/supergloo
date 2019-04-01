package utils

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/go-utils/errors"
)

const (
	// Range of a Duration in seconds, as specified in
	// google/protobuf/duration.proto. This is about 10,000 years in seconds.
	maxSeconds = int64(10000 * 365.25 * 24 * 60 * 60)
	minSeconds = -maxSeconds
)

// validateDuration determines whether the Duration is valid according to the
// definition in google/protobuf/duration.proto. A valid Duration
// may still be too large to fit into a time.Duration (the range of Duration
// is about 10,000 years, and the range of time.Duration is about 290).
func validateDuration(d *types.Duration) error {
	if d == nil {
		return errors.Errorf("duration: nil Duration")
	}
	if d.Seconds < minSeconds || d.Seconds > maxSeconds {
		return fmt.Errorf("duration: %#v: seconds out of range", d)
	}
	if d.Nanos <= -1e9 || d.Nanos >= 1e9 {
		return fmt.Errorf("duration: %#v: nanos out of range", d)
	}
	// Seconds and Nanos must have the same sign, unless d.Nanos is zero.
	if (d.Seconds < 0 && d.Nanos > 0) || (d.Seconds > 0 && d.Nanos < 0) {
		return fmt.Errorf("duration: %#v: seconds and nanos have different signs", d)
	}
	return nil
}

// DurationFromProto converts a Duration to a time.Duration. DurationFromProto
// returns an error if the Duration is invalid or is too large to be
// represented in a time.Duration.
func DurationFromProto(p *types.Duration) (*time.Duration, error) {
	if err := validateDuration(p); err != nil {
		return nil, err
	}
	d := time.Duration(p.Seconds) * time.Second
	if int64(d/time.Second) != p.Seconds {
		return nil, fmt.Errorf("duration: %#v is out of range for time.Duration", p)
	}
	if p.Nanos != 0 {
		d += time.Duration(p.Nanos)
		if (d < 0) != (p.Nanos < 0) {
			return nil, fmt.Errorf("duration: %#v is out of range for time.Duration", p)
		}
	}
	return &d, nil
}

// DurationProto converts a time.Duration to a Duration.
func DurationProto(d time.Duration) *types.Duration {
	nanos := d.Nanoseconds()
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return &types.Duration{
		Seconds: secs,
		Nanos:   int32(nanos),
	}
}
