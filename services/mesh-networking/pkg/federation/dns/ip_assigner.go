package dns

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	IpRecordName = "cluster-ip-record"

	// https://preliminary.istio.io/docs/setup/install/multicluster/gateways/#configure-the-example-services
	// should pull IPs from this network
	// TODO this should probably be configurable https://github.com/solo-io/service-mesh-hub/issues/236
	Subnet = "240.0.0.0/4"

	unAssignedIp = ""
)

var (
	NetworkExhaustion = func(cidr string) error {
		return eris.Errorf("All IPs in network %s are exhausted", cidr)
	}
	NoIpsRecordedYet = func(clusterName string) error {
		return eris.Errorf("Cluster %s has not had any IPs recorded yet", clusterName)
	}
	IpNotAMemberOfSubnet = func(ip, cidr string) error {
		return eris.Errorf("IP %s is not a member of network %s", ip, cidr)
	}
	FailedToParseCidr = func(err error, cidr string) error {
		return eris.Wrapf(err, "Failed to parse CIDR representation of subnet: %s", cidr)
	}
	UnexpectedConfigMapFormat = func(err error, mapEntry string) error {
		return eris.Wrapf(err, "Data in config map cannot be marshaled to []string: %s", mapEntry)
	}
	UnexpectedInternalError = func(err error) error {
		return eris.Wrapf(err, "Internal error while assigning new IP; this shouldn't happen")
	}

	// exponential backoff retry with an initial period of 0.1s for 7 iterations, which will mean a cumulative retry period of ~6s
	cmLookupOpts = []retry.Option{
		retry.Delay(time.Millisecond * 100),
		retry.Attempts(7),
		retry.DelayType(retry.BackOffDelay),
	}
)

func NewIpAssigner(configMapClient kubernetes_core.ConfigMapClient) IpAssigner {
	return &ipAssigner{
		configMapClient: configMapClient,
	}
}

type ipAssigner struct {
	configMapClient kubernetes_core.ConfigMapClient
}

func (i *ipAssigner) AssignIPOnCluster(ctx context.Context, clusterName string) (string, error) {
	ipRecordRef := &core_types.ResourceRef{
		Name:      IpRecordName,
		Namespace: env.GetWriteNamespace(),
	}

	ipRecordConfigMap, err := i.configMapClient.Get(ctx, clients.ResourceRefToObjectKey(ipRecordRef))

	if errors.IsNotFound(err) {
		newIpRecord := map[string]string{}
		newIp, err := generateNewIp(newIpRecord, clusterName)
		if err != nil {
			return "", err
		}

		return newIp, i.configMapClient.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: clients.ResourceRefToObjectMeta(ipRecordRef),
			Data:       newIpRecord,
		})
	} else if err != nil {
		return "", err
	} else {
		ipRecord := ipRecordConfigMap.Data
		if ipRecord == nil {
			ipRecord = map[string]string{}
		}

		newIp, err := generateNewIp(ipRecord, clusterName)
		if err != nil {
			return "", err
		}

		ipRecordConfigMap.Data = ipRecord

		return newIp, i.configMapClient.Update(ctx, ipRecordConfigMap)
	}
}

func (i *ipAssigner) UnAssignIPOnCluster(ctx context.Context, clusterName, ipToUnassign string) error {
	ipRecordRef := &core_types.ResourceRef{
		Name:      IpRecordName,
		Namespace: env.GetWriteNamespace(),
	}

	var ipRecordConfigMap *corev1.ConfigMap
	err := retry.Do(func() error {
		cm, err := i.configMapClient.Get(ctx, clients.ResourceRefToObjectKey(ipRecordRef))
		if err != nil {
			return err
		}

		if _, ok := cm.Data[clusterName]; !ok {
			return NoIpsRecordedYet(clusterName)
		}
		ipRecordConfigMap = cm
		return nil
	}, cmLookupOpts...)
	if err != nil {
		return err
	}

	storedIpsJson := ipRecordConfigMap.Data[clusterName]
	var storedIps []string
	err = json.Unmarshal([]byte(storedIpsJson), &storedIps)
	if err != nil {
		return UnexpectedConfigMapFormat(err, storedIpsJson)
	}

	for index, storedIp := range storedIps {
		if storedIp == ipToUnassign {
			storedIps[index] = unAssignedIp
		}
	}

	updatedIpList, _ := json.Marshal(storedIps)
	ipRecordConfigMap.Data[clusterName] = string(updatedIpList)

	return i.configMapClient.Update(ctx, ipRecordConfigMap)
}

func generateNewIp(cmData map[string]string, clusterName string) (string, error) {
	recordJson, ok := cmData[clusterName]

	// if we have already stored entries, use those
	var alreadyAssignedIps []string
	if ok {
		err := json.Unmarshal([]byte(recordJson), &alreadyAssignedIps)
		if err != nil {
			return "", UnexpectedConfigMapFormat(err, recordJson)
		}
	}

	// the IP that we have most recently assigned
	// defaults to the base IP
	lastAssignedIp, _, err := net.ParseCIDR(Subnet)
	if err != nil {
		return "", UnexpectedInternalError(err)
	}
	lastAssignedIpString := lastAssignedIp.String()
	lastAssignedIpIndex := -1

	// go through our assigned entries until we find a hole to fill
	// when we un-assign IPs, we set them to "" in the array
	// if we have not assigned any yet, then we will increment the base IP (ie, 240.0.0.0)
	for index, ip := range alreadyAssignedIps {
		if ip == unAssignedIp {
			break
		}

		lastAssignedIpString = ip
		lastAssignedIpIndex = index
	}

	nextIp, err := incrementIpInNetwork(lastAssignedIpString, Subnet)
	if err != nil {
		return "", err
	}

	// if we are filling a hole, insert the IP at that hole
	// otherwise, append it to the end of the list
	if len(alreadyAssignedIps) > 0 && lastAssignedIpIndex >= -1 && lastAssignedIpIndex < len(alreadyAssignedIps)-1 {
		alreadyAssignedIps[lastAssignedIpIndex+1] = nextIp
	} else {
		alreadyAssignedIps = append(alreadyAssignedIps, nextIp)
	}

	// throw away the error, as we know we're working with a []string and that's definitely going to .Marshal()
	newIpList, _ := json.Marshal(alreadyAssignedIps)
	cmData[clusterName] = string(newIpList)

	return nextIp, nil
}

// given an IP and a CIDR representation of the subnet it is a member of,
// produce the next valid IP. An error is returned if the subnet IP
// space is exhausted.
func incrementIpInNetwork(origIPString, cidr string) (string, error) {
	originalIp := net.ParseIP(origIPString)
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", FailedToParseCidr(err, cidr)
	}

	// if the subnet we're handed doesn't even contain the original IP, bail out
	if !ipNet.Contains(originalIp) {
		return "", IpNotAMemberOfSubnet(origIPString, cidr)
	}

	// an IP is four octets, and is represented in Go's "net" package as a `[4]byte`
	// starting with the "right-most" octet, the least significant bits, increment
	// the bytes (potentially carrying to the next octet)
	for i := len(originalIp) - 1; i >= 0; i-- {
		originalIp[i]++
		if originalIp[i] != 0 {
			break
		}
	}

	// if we have walked out of the subnet space, that is an error
	if !ipNet.Contains(originalIp) {
		return "", NetworkExhaustion(cidr)
	}
	return originalIp.String(), nil
}
