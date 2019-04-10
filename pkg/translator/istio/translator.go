package istio

import (
	"context"
	"reflect"
	"sort"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type Translator interface {
	// translates a snapshot into a set of istio configs for each mesh
	Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error)
}

// A container for the entire set of config for a single istio mesh
type MeshConfig struct {
	// mtls
	MeshPolicy *v1alpha1.MeshPolicy // meshpolicy is a singleton

	// root cert
	RootCert *v1.TlsSecret // singleton

	// routing
	DestinationRules v1alpha3.DestinationRuleList
	VirtualServices  v1alpha3.VirtualServiceList

	// rbac
	SecurityConfig
}

func (c *MeshConfig) Sort() {
	sort.SliceStable(c.DestinationRules, func(i, j int) bool {
		return c.DestinationRules[i].Metadata.Less(c.DestinationRules[j].Metadata)
	})
	sort.SliceStable(c.VirtualServices, func(i, j int) bool {
		return c.VirtualServices[i].Metadata.Less(c.VirtualServices[j].Metadata)
	})
}

// first create all destination rules for all subsets of each upstream
// then we need to apply the ISTIO_MUTUAL policy depending on
// whether mtls is enabled

type translator struct {
	plugins []plugins.Plugin
}

func NewTranslator(plugins []plugins.Plugin) Translator {
	return &translator{plugins: plugins}
}

/*
Translate a snapshot into a set of MeshConfigs for each mesh
Currently only active istio mesh is expected.
*/
func (t *translator) Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error) {
	meshes := snapshot.Meshes.List()
	meshGroups := snapshot.Meshgroups.List()
	upstreams := snapshot.Upstreams.List()
	pods := snapshot.Pods.List()
	routingRules := snapshot.Routingrules.List()
	securityRules := snapshot.Securityrules.List()

	resourceErrs := make(reporter.ResourceErrors)
	resourceErrs.Accept(meshes.AsInputResources()...)
	resourceErrs.Accept(meshGroups.AsInputResources()...)
	resourceErrs.Accept(routingRules.AsInputResources()...)

	utils.ValidateMeshGroups(meshes, meshGroups, resourceErrs)

	routingRulesByMesh := utils.GroupRulesByMesh(routingRules, securityRules, meshes, meshGroups, resourceErrs)

	perMeshConfig := make(map[*v1.Mesh]*MeshConfig)

	params := plugins.Params{
		Ctx:       ctx,
		Upstreams: upstreams,
	}

	tlsSecrets := snapshot.Tlssecrets.List()

	for _, mesh := range meshes {
		istio, ok := mesh.MeshType.(*v1.Mesh_Istio)
		if !ok {
			// we only want istio meshes
			continue
		}
		writeNamespace := istio.Istio.InstallationNamespace
		rules := routingRulesByMesh[mesh]
		in := inputMeshConfig{
			writeNamespace: writeNamespace,
			mesh:           mesh,
			rules:          rules,
		}
		meshConfig, err := t.translateMesh(params, in, upstreams, tlsSecrets, pods, resourceErrs)
		if err != nil {
			resourceErrs.AddError(mesh, errors.Wrapf(err, "translating mesh config"))
			contextutils.LoggerFrom(ctx).Errorf("translating for mesh %v failed: %v", mesh.Metadata.Ref(), err)
			continue
		}
		perMeshConfig[mesh] = meshConfig
	}

	return perMeshConfig, resourceErrs, nil
}

type inputMeshConfig struct {
	// where crds should be written. this is normally the mesh installation namespace
	writeNamespace string
	// the mesh we're configuring
	mesh *v1.Mesh
	// list of rules which apply to this mesh
	rules utils.RuleSet
}

// produces a complete istio config
func (t *translator) translateMesh(
	params plugins.Params,
	input inputMeshConfig,
	upstreams gloov1.UpstreamList,
	tlsSecrets v1.TlsSecretList,
	pods v1.PodList,
	resourceErrs reporter.ResourceErrors) (*MeshConfig, error) {
	ctx := params.Ctx
	mtlsEnabled := input.mesh.MtlsConfig != nil && input.mesh.MtlsConfig.MtlsEnabled
	rules := input.rules

	destinationHostsPortsAndLabels, err := utils.LabelsAndPortsByHost(upstreams)
	if err != nil {
		return nil, errors.Wrapf(err, "internal error: getting ports and labels from upstreams")
	}
	var destinationRules v1alpha3.DestinationRuleList
	var virtualServices v1alpha3.VirtualServiceList
	for destinationHost, destinationPortAndLabelSets := range destinationHostsPortsAndLabels {
		var labelSets []map[string]string
		// must find unique label sets; they will be repeated for multiple ports
	findUniqueSets:
		for _, set := range destinationPortAndLabelSets {
			for _, existing := range labelSets {
				if reflect.DeepEqual(set.Labels, existing) {
					continue findUniqueSets
				}
			}
			labelSets = append(labelSets, set.Labels)
		}

		dr := makeDestinationRule(ctx,
			input.writeNamespace,
			destinationHost,
			labelSets,
			mtlsEnabled,
		)
		destinationRules = append(destinationRules, dr)

		vs := t.makeVirtualServiceForHost(ctx,
			params,
			input.writeNamespace,
			destinationHost,
			destinationPortAndLabelSets,
			rules.Routing,
			upstreams,
			resourceErrs,
		)

		virtualServices = append(virtualServices, vs)
	}

	var meshPolicy *v1alpha1.MeshPolicy
	if mtlsEnabled {
		meshPolicy = &v1alpha1.MeshPolicy{
			Metadata: core.Metadata{
				// the required name for istio MeshPolicy
				// https://istio.io/docs/tasks/security/authn-policy/#globally-enabling-istio-mutual-tls
				Name: "default",
			},
			Peers: []*v1alpha1.PeerAuthenticationMethod{{
				Params: &v1alpha1.PeerAuthenticationMethod_Mtls{Mtls: &v1alpha1.MutualTls{
					Mode: v1alpha1.MutualTls_STRICT,
				}},
			}},
		}
	}

	var rootCert *v1.TlsSecret
	if input.mesh.MtlsConfig != nil && input.mesh.MtlsConfig.RootCertificate != nil {
		tlsSecret, err := tlsSecrets.Find(input.mesh.MtlsConfig.RootCertificate.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "finding tls secret for mesh root cert")
		}
		// set cacerts secret for istio
		// https://istio.io/docs/tasks/security/plugin-ca-cert/#plugging-in-the-existing-certificate-and-key
		rootCert = &v1.TlsSecret{
			Metadata: core.Metadata{
				Namespace: input.writeNamespace,
				Name:      "cacerts",
			},
			RootCert:  tlsSecret.RootCert,
			CertChain: tlsSecret.CertChain,
			CaCert:    tlsSecret.CaCert,
			CaKey:     tlsSecret.CaKey,
		}
	}

	securityConfig := createSecurityConfig(
		input.writeNamespace,
		input.rules.Security,
		upstreams,
		pods,
		resourceErrs,
	)

	meshConfig := &MeshConfig{
		VirtualServices:  virtualServices,
		DestinationRules: destinationRules,
		MeshPolicy:       meshPolicy,
		SecurityConfig:   securityConfig,
		RootCert:         rootCert,
	}
	meshConfig.Sort()

	return meshConfig, nil
}
