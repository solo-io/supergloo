package meshserviceutils

import (
	"hash/fnv"
	"net"

	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// We must generate IPs to use for Service Entries. for now we simply generate them from this subnet.
// https://preliminary.istio.io/docs/setup/install/multicluster/gateways/#configure-the-example-services
//
// TODO(ilackarms): allow this to be inferred, configured by the user, or remove when
// istio supports creating service entries without assigning IPs.
var ipAssignableSubnet = "240.0.0.0/4"

const (
	kubeService     = "kube-service"
	failoverService = "failover-service"
)

func ConstructUniqueIpForKubeService(kubeServiceRef ezkube.ClusterResourceId) (net.IP, error) {
	return constructUniqueIp(kubeServiceRef, kubeService)
}

func ConstructUniqueIpForFailoverService(failoverServiceRef ezkube.ResourceId) (net.IP, error) {
	return constructUniqueIp(&v1.ClusterObjectRef{
		Name:        failoverServiceRef.GetName(),
		Namespace:   failoverServiceRef.GetNamespace(),
		ClusterName: "",
	}, failoverService)
}

func constructUniqueIp(clusterObjectRef ezkube.ClusterResourceId, scope string) (net.IP, error) {
	ip, cidr, err := net.ParseCIDR(ipAssignableSubnet)
	if err != nil {
		return nil, err
	}
	ip = ip.Mask(cidr.Mask)
	if len(ip) != 4 {
		return nil, eris.Errorf("unexpected length for cidr IP: %v", len(ip))
	}

	h := fnv.New32()
	if _, err := h.Write([]byte(clusterObjectRef.GetName())); err != nil {
		return nil, err
	}
	if _, err := h.Write([]byte(clusterObjectRef.GetNamespace())); err != nil {
		return nil, err
	}
	if _, err := h.Write([]byte(clusterObjectRef.GetClusterName())); err != nil {
		return nil, err
	}
	if _, err := h.Write([]byte(scope)); err != nil {
		return nil, err
	}
	hash := h.Sum32()
	var hashedIP net.IP = []byte{
		byte(hash),
		byte(hash >> 8),
		byte(hash >> 16),
		byte(hash >> 24),
	}
	hashedIP.Mask(cidr.Mask)

	for i := range hashedIP {
		hashedIP[i] = hashedIP[i] | ip[i]
	}

	return hashedIP, nil
}
