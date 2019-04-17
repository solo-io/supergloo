package surveyutils

import (
	"github.com/solo-io/go-utils/surveyutils"
	"github.com/solo-io/supergloo/cli/pkg/constants"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func SurveyMaxRetries(opts *options.MaxRetries) error {
	if err := surveyutils.GetUint32Input(flagutils.Description_MaxRetries_Attempts, &opts.Attempts); err != nil {
		return err
	}
	if err := surveyutils.GetDurationInput(flagutils.Description_MaxRetries_PerTryTimeout, &opts.PerTryTimeout); err != nil {
		return err
	}
	if err := surveyutils.ChooseFromList(flagutils.Description_MaxRetries_RetryOn, &opts.RetryOn, constants.PossibleMaxRetry_RetryOnValues); err != nil {
		return err
	}
	return nil
}

func SurveyRetryBudget(opts *v1.RetryBudget) error {
	if err := surveyutils.GetFloat32Input(flagutils.Description_RetryBudget_RetryRatio, &opts.RetryRatio); err != nil {
		return err
	}
	if err := surveyutils.GetUint32Input(flagutils.Description_RetryBudget_MinRetriesPerSecond, &opts.MinRetriesPerSecond); err != nil {
		return err
	}
	if err := surveyutils.GetDurationInput(flagutils.Description_RetryBudget_Ttl, &opts.Ttl); err != nil {
		return err
	}
	return nil
}
