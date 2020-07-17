package destinationrule

/*
This file contains the registered set of plugins for the translator.

A plugin must be imported into this file in order to be used by the translator
*/

import (
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/destinationrule/plugins/outlierdetection"
)
