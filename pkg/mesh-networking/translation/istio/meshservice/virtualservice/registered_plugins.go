package virtualservice

/*
This file contains the registered set of plugins for the translator.

A plugin must be imported into this file in order to be used by the translator
*/

import (
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins/cors"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins/faultinjection"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins/mirror"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins/retries"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins/timeout"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins/trafficshift"
)
