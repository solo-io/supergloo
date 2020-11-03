package defaults

var (
	// "istio": "ingressgateway" is a known string pair to Istio- it's semantically meaningful but unfortunately not exported from anywhere
	// their ingress gateway is hardcoded in their own implementation to have this label
	// https://github.com/istio/istio/blob/4e27ddc64f6a12e622c4cd5c836f5d7edf94e971/istioctl/cmd/describe.go#L1138
	DefaultIngressGatewayWorkloadLabels = map[string]string{
		"istio": "ingressgateway",
	}
	// "istio": "egressgateway" is a known string pair to Istio- it's semantically meaningful but unfortunately not exported from anywhere
	// their egress gateway is hardcoded in their own implementation to have this label
	DefaultEgressGatewayWorkloadLabels = map[string]string{
		"istio": "egressgateway",
	}
)
