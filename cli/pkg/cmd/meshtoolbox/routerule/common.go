package routerule

import (
	"fmt"
	"strconv"
	"strings"

	types "github.com/gogo/protobuf/types"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common/iutil"
)

func EnsureDuration(rootMessage string, durOpts *options.InputDuration, targetDur *types.Duration, opts *options.Options) error {
	dur := types.Duration{}
	if !opts.Top.Static {
		err := iutil.GetStringInput(fmt.Sprintf("%v (seconds)", rootMessage), &durOpts.Seconds)
		if err != nil {
			return err
		}
		err = iutil.GetStringInput(fmt.Sprintf("%v (nanoseconds)", rootMessage), &durOpts.Nanos)
		if err != nil {
			return err
		}
	}
	// if not in interactive mode, timeout values will have already been passed
	if durOpts.Seconds != "" {
		sec, err := strconv.Atoi(durOpts.Seconds)
		if err != nil {
			return err
		}
		dur.Seconds = int64(sec)
	}
	if durOpts.Nanos != "" {
		nanos, err := strconv.Atoi(durOpts.Nanos)
		if err != nil {
			return err
		}
		dur.Nanos = int32(nanos)
	}
	*targetDur = dur
	return nil
}

// EnsurePercentage transforms a source string to a target int
// If not present, it promts the user for input with the given message
// Errors on invalid input
func EnsurePercentage(message string, source *string, target *int32, opts *options.Options) error {
	if !opts.Top.Static {
		if err := iutil.GetStringInput(message, source); err != nil {
			return err
		}
	}
	if *source != "" {
		percentage, err := strconv.Atoi(*source)
		if err != nil {
			return err
		}
		*target = int32(percentage)
	}
	return nil
}

func ensureCsv(message string, source string, target *[]string, staticMode bool) error {
	if staticMode && source == "" {
		return fmt.Errorf(message)
	}
	if !staticMode {
		if err := iutil.GetStringInput(message, &source); err != nil {
			return err
		}
	}
	parts := strings.Split(source, ",")
	*target = parts
	return nil
}

// Expected format of source: k1,v1,k2,v2
func ensureKVCsv(message string, source string, target *map[string]string, staticMode bool) error {
	parts := []string{}
	ensureCsv(message, source, &parts, staticMode)
	if len(parts)%2 != 0 {
		return fmt.Errorf("Must provide one key per value (received an odd sum)")
	}
	for i := 0; i < len(parts)/2; i++ {
		(*target)[parts[i*2]] = parts[i*2+1]
	}
	return nil
}
