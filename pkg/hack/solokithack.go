package hack

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

type ClusterWatchAggregator struct {
	aggregator wrapper.WatchAggregator
	watchers   map[string][]clients.ResourceWatcher
}

var _ multicluster.ClientForClusterHandler = &ClusterWatchAggregator{}
var _ handler.ClusterHandler = &ClusterWatchAggregator{}

// Provides a ClientForClusterHandler to collect all watchers available on a cluster.
// Provides a ClusterHandler to sync all clients to the aggregated watch.
// NOTE that all clients must be added to the ClientForClusterHandler before ANY are sent to the watch.
func NewAggregatedWatchClusterClientHandler(aggregator wrapper.WatchAggregator) *ClusterWatchAggregator {
	return &ClusterWatchAggregator{
		aggregator: aggregator,
		watchers:   make(map[string][]clients.ResourceWatcher),
	}
}

func (h *ClusterWatchAggregator) HandleNewClusterClient(cluster string, client clients.ResourceClient) {
	h.watchers[cluster] = append(h.watchers[cluster], client)
}

func (h *ClusterWatchAggregator) HandleRemovedClusterClient(cluster string, client clients.ResourceClient) {
	// noop
}

func (h *ClusterWatchAggregator) ClusterAdded(cluster string, restConfig *rest.Config) {
	for _, w := range h.watchers[cluster] {
		h.aggregator.AddWatch(w)
	}
}

func (h *ClusterWatchAggregator) ClusterRemoved(cluster string, restConfig *rest.Config) {
	for _, w := range h.watchers[cluster] {
		h.aggregator.RemoveWatch(w)
	}
	delete(h.watchers, cluster)
}
