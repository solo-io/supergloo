package surveyutils

import (
	"fmt"
	"strings"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/go-utils/errors"
)

func SurveyMapStringString(in *map[string]string) error {
	for {
		var kvPair string
		if err := cliutil.GetStringInput("enter a key-value pair in the format KEY=VAL. "+
			"leave empty to finish", &kvPair); err != nil {
			return err
		}
		if kvPair == "" {
			return nil
		}
		split := strings.SplitN(kvPair, "=", 2)
		if len(split) != 2 {
			return errors.Errorf("key-value pair must be in the format KEY=VAL, you entered %v", kvPair)
		}
		m := *in
		m[split[0]] = split[1]
		*in = m
		fmt.Printf("%v\n", *in)
	}
}
