package plugins

/*
This file contains the registered set of plugins for the translator.

A plugin must be imported into this file in order to be used by the translator
*/

import (
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/cors"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/faultinjection"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/mirror"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/outlierdetection"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/retries"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/timeout"
	_ "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/trafficshift"
)
