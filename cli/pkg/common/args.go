package common

import (
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/spf13/cobra"
)

func RequiredNameArg(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.Errorf("name argument is required, 0 args found")
	}
	return nil
}
