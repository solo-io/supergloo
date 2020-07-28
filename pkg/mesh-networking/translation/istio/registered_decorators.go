package istio

/*
This file contains the registered set of decorators for the translator.

A decorator must be imported into this file in order to be used by the translator
*/

import (
	// TrafficPolicy decorators
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/cors"
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/faultinjection"
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/headermanipulation"
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/mirror"
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/outlierdetection"
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/retries"
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/timeout"
	_ "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/trafficshift"
)
