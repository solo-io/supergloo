package routerule

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	API_CONFIG = "api_config"
	CLI_CONFIG = "cli_config"
)

func configFromFile(opts *options.Options) (string, error) {
	//opts.Create.InputRoutingRule = options.InputRoutingRule{}
	cliInput := &options.InputRoutingRule{}
	apiInput := &v1.RoutingRule{}
	if err := genericReadFileInto(opts.Top.File, apiInput); err == nil {
		(*opts).MeshTool.RoutingRule = *apiInput
		return API_CONFIG, nil
	}
	if err := genericReadFileInto(opts.Top.File, cliInput); err == nil {
		(*opts).Create.InputRoutingRule = *cliInput
		return CLI_CONFIG, nil
	}
	return "", errors.Errorf("yaml file does not conform to either cli format, or API format")
}

func genericReadFileInto(filename string, dat interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Errorf("error reading file: %v", err)
	}
	jsn, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsn, dat)
}
