package consul

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/consul"

	"github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	metadataName = "supergloo.name"
)

type PolicySyncer struct {
}

func (s *PolicySyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "consul-policy-syncer")

	for _, mesh := range snap.Meshes.List() {
		_, ok := mesh.MeshType.(*v1.Mesh_Consul)
		if !ok {
			// not our mesh, we don't care
			continue
		}

		policy := mesh.Policy
		if policy != nil {
			if err := s.syncPolicy(ctx, snap.Upstreams, policy); err != nil {
				return err
			}
		}
	}
	return nil
}

func get(upstreams gloov1.UpstreamsByNamespace, ref core.ResourceRef) (*consul.UpstreamSpec, error) {
	up, err := upstreams[ref.Namespace].Find(ref.Namespace, ref.Name)
	if err != nil {
		return nil, err
	}
	spec, ok := up.UpstreamSpec.UpstreamType.(*gloov1.UpstreamSpec_Consul)
	if !ok {
		return nil, errors.New("not consul upstream")
	}
	return spec.Consul, nil
}

func (s *PolicySyncer) syncPolicy(ctx context.Context, upstreams gloov1.UpstreamsByNamespace, p *v1.Policy) error {
	logger := contextutils.LoggerFrom(ctx)

	var client *api.Client
	var err error
	client, err = api.NewClient(api.DefaultConfig())

	if err != nil {
		return err
	}
	connectClient := client.Connect()

	// create desired intentions
	var desiredIntentions []*api.Intention
	for _, rule := range p.Rules {
		if rule.Source == nil {
			// TODO: should we return error instead?
			continue
		}
		if rule.Destination == nil {
			// TODO: should we return error instead?
			continue
		}

		consuleSource, err := get(upstreams, *rule.Source)
		if err != nil {
			return err
		}

		consuleDestination, err := get(upstreams, *rule.Destination)
		if err != nil {
			return err
		}

		// need to convert our upstream to a consul service?
		name := fmt.Sprintf("consuleSource.ServiceName-consuleDestination.ServiceName")
		desiredIntentions = append(desiredIntentions, &api.Intention{
			Action:          api.IntentionActionAllow,
			Meta:            map[string]string{metadataName: name},
			SourceName:      consuleSource.ServiceName,
			DestinationName: consuleDestination.ServiceName,
		})
	}
	// create an intention and hope for the best!
	// get all intentions
	intentions, _, err := connectClient.Intentions(nil)
	if err != nil {
		logger.Warnw("error getting intentions", "err", err)
		return err
	}

	// find intentions to remove
	var removeThese []*api.Intention
Outloop:
	for _, intention := range intentions {
		if intention.Meta == nil {
			continue
		}
		if len(intention.Meta[metadataName]) == 0 {
			continue
		}
		// this intention is own by us, let's see if it's still needed
		for i, desiredIntention := range desiredIntentions {
			if intention.Meta[metadataName] == desiredIntention.Meta[metadataName] {
				// this is desired exists. remove it from desired as we don't need to ad it.
				desiredIntentions = append(desiredIntentions[:i], desiredIntentions[i+1:]...)
				continue Outloop
			}
		}
		// if we got here, it means that this intention is not desired
		removeThese = append(removeThese, intention)
	}

	logger.Infow("Adding intentions", "toadd", desiredIntentions, "toremove", removeThese)

	var multiErr *multierror.Error
	for _, intention := range desiredIntentions {
		_, _, err := connectClient.IntentionCreate(intention, nil)
		if err != nil {
			logger.Warnw("error adding intention", "intention", intention, "err", err)
			multiErr = multierror.Append(multiErr, err)
		}

	}

	for _, intention := range removeThese {
		_, err := connectClient.IntentionDelete(intention.ID, nil)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr.ErrorOrNil()
}
