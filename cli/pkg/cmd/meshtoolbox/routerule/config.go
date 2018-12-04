package routerule

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"io/ioutil"
)



func configFromFile(opts *options.Options) error {
	//opts.Create.InputRoutingRule = options.InputRoutingRule{}
	fileInput := &options.InputRoutingRule{}
	err := genericReadFileInto(opts.Top.File, fileInput)
	if err != nil {
		return err
	}
	// set daa
	(*opts).Create.InputRoutingRule = *fileInput
	return nil
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
