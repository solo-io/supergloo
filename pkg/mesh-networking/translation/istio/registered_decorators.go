package istio

/*
This file contains the registered set of decorators for the translator.

A decorator must be imported into this file in order to be used by the translator
*/

import (
	// TrafficPolicy decorators
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/cors"
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/faultinjection"
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/headermanipulation"
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mirror"
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/outlierdetection"
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/retries"
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/timeout"
	_ "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/trafficshift"
)
