package surveyutils

import (
	"strconv"
	"time"

	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveyMaxRetries(opts *options.MaxRetries) error {
	if err := cliutil.GetUint32Input(flagutils.Description_MaxRetries_Attempts, &opts.Attempts); err != nil {
		return err
	}
	if err := GetDurationInput(flagutils.Description_MaxRetries_PerTryTimeout, &opts.PerTryTimeout); err != nil {
		return err
	}
	if err := cliutil.GetStringInput(flagutils.Description_MaxRetries_RetryOn, &opts.RetryOn); err != nil {
		return err
	}
	return nil
}

func SurveyRetryBudget(opts *v1.RetryBudget) error {
	if err := GetFloat32Input(flagutils.Description_RetryBudget_RetryRatio, &opts.RetryRatio); err != nil {
		return err
	}
	if err := cliutil.GetUint32Input(flagutils.Description_RetryBudget_MinRetriesPerSecond, &opts.MinRetriesPerSecond); err != nil {
		return err
	}
	if err := GetDurationInput(flagutils.Description_RetryBudget_Ttl, &opts.Ttl); err != nil {
		return err
	}
	return nil
}

// TODO: move these 2 funcs to go-utils
func GetDurationInput(msg string, duration *time.Duration) error {
	var durStr string
	if err := cliutil.GetStringInput(msg, &durStr); err != nil {
		return err
	}
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return err
	}
	*duration = dur
	return nil
}
func GetFloat32Input(msg string, value *float32) error {
	var strValue string
	prompt := &survey.Input{Message: msg, Default: "0"}
	if err := cliutil.AskOne(prompt, &strValue, nil); err != nil {
		return err
	}
	val, err := strconv.ParseFloat(strValue, 32)
	if err != nil {
		return err
	}
	*value = float32(val)
	return nil
}
