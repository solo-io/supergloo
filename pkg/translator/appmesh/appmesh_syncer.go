package istio

import (
	"context"
	"sort"
	"sync"
	"unicode/utf8"

	"github.com/solo-io/supergloo/pkg/translator/utils"

	"github.com/solo-io/solo-kit/pkg/utils/log"

	"github.com/mitchellh/hashstructure"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

// NOTE: copy-pasted from discovery/pkg/fds/discoveries/aws/aws.go
// TODO: aggregate these somewhere
const (
	// expected map identifiers for secrets
	awsAccessKey = "access_key"
	awsSecretKey = "secret_key"
)

// TODO: util method
func NewAwsClientFromSecret(awsSecret *gloov1.Secret_Aws, region string) (*appmesh.AppMesh, error) {
	accessKey := awsSecret.Aws.AccessKey
	if accessKey != "" && !utf8.Valid([]byte(accessKey)) {
		return nil, errors.Errorf("%s not a valid string", awsAccessKey)
	}
	secretKey := awsSecret.Aws.SecretKey
	if secretKey != "" && !utf8.Valid([]byte(secretKey)) {
		return nil, errors.Errorf("%s not a valid string", awsSecretKey)
	}
	sess, err := session.NewSession(aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, "")))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create AWS session")
	}
	svc := appmesh.New(sess, &aws.Config{Region: aws.String(region)})

	return svc, nil
}

// todo: replace with interface
type AppMeshClient = *appmesh.AppMesh

type MeshRoutingSyncer struct {
	lock           sync.Mutex
	activeSessions map[uint64]AppMeshClient
}

func NewMeshRoutingSyncer() *MeshRoutingSyncer {
	return &MeshRoutingSyncer{
		lock:           sync.Mutex{},
		activeSessions: make(map[uint64]AppMeshClient),
	}
}

func hashCredentials(awsSecret *gloov1.Secret_Aws, region string) uint64 {
	hash, _ := hashstructure.Hash(struct {
		awsSecret *gloov1.Secret_Aws
		region    string
	}{
		awsSecret: awsSecret,
		region:    region,
	}, nil)
	return hash
}

func (s *MeshRoutingSyncer) NewOrCachedClient(appMesh *v1.AppMesh, secrets gloov1.SecretList) (AppMeshClient, error) {
	secret, err := secrets.Find(appMesh.AwsCredentials.Strings())
	if err != nil {
		return nil, errors.Wrapf(err, "finding aws credentials for mesh")
	}
	region := appMesh.AwsRegion
	if region == "" {
		return nil, errors.Wrapf(err, "mesh must provide aws_region")
	}

	awsSecret, ok := secret.Kind.(*gloov1.Secret_Aws)
	if !ok {
		return nil, errors.Errorf("mesh referenced non-AWS secret, AWS secret required")
	}
	if awsSecret.Aws == nil {
		return nil, errors.Errorf("secret missing field Aws")
	}

	// check if we already have an active session for this region/credential
	sessionKey := hashCredentials(awsSecret, region)
	s.lock.Lock()
	appMeshClient, ok := s.activeSessions[sessionKey]
	s.lock.Unlock()
	if !ok {
		// create a new client and cache it
		// TODO: is there a point where we should drop old sessions?
		// maybe aws will do it for us
		appMeshClient, err = NewAwsClientFromSecret(awsSecret, region)
		if err != nil {
			return nil, errors.Wrapf(err, "creating aws client from provided secret/region")
		}
		s.lock.Lock()
		s.activeSessions[sessionKey] = appMeshClient
		s.lock.Unlock()
	}

	return appMeshClient, nil
}

func (s *MeshRoutingSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "mesh-routing-syncer")
	logger := contextutils.LoggerFrom(ctx)
	meshes := snap.Meshes.List()
	upstreams := snap.Upstreams.List()
	rules := snap.Routingrules.List()

	logger.Infof("begin sync %v (%v meshes, %v upstreams, %v rules)", snap.Hash(),
		len(meshes), len(upstreams), len(rules))
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("%v", snap)

	for _, mesh := range snap.Meshes.List() {
		if err := s.sync(mesh, snap); err != nil {
			return errors.Wrapf(err, "syncing mesh %v", mesh.Metadata.Ref())
		}
	}

	return nil
	/*
		0 - mesh per mesh
		1 - virtual node per upstream
		2 - routing rules get aggregated into virtual service like object
		routes on virtual service become the aws routes
		virtual service becomes virtual router
	*/
	//exampleMesh := appmesh.CreateMeshInput{}
	//exampleVirtualNode := appmesh.CreateVirtualNodeInput{}
	//exampleVirtualRouter := appmesh.CreateVirtualRouterInput{}
	//exampleRoute := appmesh.CreateRouteInput{}
	//
	//destinationRules, err := virtualNodesForUpstreams(rules, meshes, upstreams)
	//if err != nil {
	//	return errors.Wrapf(err, "creating subsets from snapshot")
	//}
	//
	//virtualServices, err := virtualServicesForRules(rules, meshes, upstreams)
	//if err != nil {
	//	return errors.Wrapf(err, "creating virtual services from snapshot")
	//}
	//return s.writeIstioCrds(ctx, destinationRules, virtualServices)
}

func (s *MeshRoutingSyncer) sync(mesh *v1.Mesh, snap *v1.TranslatorSnapshot) error {
	appMesh, ok := mesh.MeshType.(*v1.Mesh_AppMesh)
	if !ok {
		return nil
	}
	if appMesh.AppMesh == nil {
		return errors.Errorf("%v missing configuration for AppMesh", mesh.Metadata.Ref())
	}
	desiredMesh := &appmesh.CreateMeshInput{
		MeshName: aws.String(mesh.Metadata.Name),
	}
	upstreams := snap.Upstreams.List()
	virtualNodes, err := virtualNodesFromUpstreams(upstreams)
	if err != nil {
		return errors.Wrapf(err, "creating virtual nodes from upstreams")
	}
	virtualRouters, err := virtualRoutersFromUpstreams(upstreams)
	if err != nil {
		return errors.Wrapf(err, "creating virtual routers from upstreams")
	}
	//routes := routesForRules()

	// todo: generate desired virtual node for every service
	// go through route rules, only focus on traffic shifting, otherwise notihng to do?
	// * need to investigate using testrunner pod with the yaml
}

func (s *MeshRoutingSyncer) reconcileMesh(mesh appmesh.CreateMeshInput,
	vNodes []appmesh.CreateVirtualNodeInput,
	vRouters []appmesh.CreateVirtualRouterInput,
	vRoutes []appmesh.CreateRouteInput) error {
	// todo: look up the right client

}

func virtualNodesFromUpstreams(meshName string, list gloov1.UpstreamList) ([]appmesh.CreateVirtualNodeInput, error) {
	portsByHost := make(map[string][]uint32)
	// TODO: filter hosts by policy, i.e. only what the user wants
	var allHosts []string
	for _, us := range list {
		host, err := utils.GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		port, err := utils.GetPortForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting port for upstream")
		}
		portsByHost[host] = append(portsByHost[host], port)
		allHosts = append(allHosts, host)
	}

	var virtualNodes []appmesh.CreateVirtualNodeInput
	for host, ports := range portsByHost {
		var listeners []*appmesh.Listener
		for _, port := range ports {
			listener := &appmesh.Listener{
				PortMapping: &appmesh.PortMapping{
					// TODO: support more than just http here
					Protocol: aws.String("http"),
					Port:     aws.Int64(int64(port)),
				},
			}
			listeners = append(listeners, listener)
		}
		virtualNode := appmesh.CreateVirtualNodeInput{
			MeshName: aws.String(meshName),
			Spec: &appmesh.VirtualNodeSpec{
				Backends:  aws.StringSlice(allHosts),
				Listeners: listeners,
				ServiceDiscovery: &appmesh.ServiceDiscovery{
					Dns: &appmesh.DnsServiceDiscovery{
						ServiceName: aws.String(host),
					},
				},
			},
		}
		virtualNodes = append(virtualNodes, virtualNode)
	}
	sort.SliceStable(virtualNodes, func(i, j int) bool {
		return virtualNodes[i].String() < virtualNodes[j].String()
	})
	return virtualNodes, nil
}

func virtualRoutersFromUpstreams(meshName string, list gloov1.UpstreamList) ([]appmesh.CreateVirtualRouterInput, error) {
	var allHosts []string
	for _, us := range list {
		host, err := utils.GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		allHosts = append(allHosts, host)
	}

	var virtualRouters []appmesh.CreateVirtualRouterInput
	for _, host := range allHosts {
		virtualRouter := appmesh.CreateVirtualRouterInput{
			MeshName:          aws.String(meshName),
			VirtualRouterName: aws.String(host),
			Spec: &appmesh.VirtualRouterSpec{
				ServiceNames: aws.StringSlice([]string{host}),
			},
		}
		virtualRouters = append(virtualRouters, virtualRouter)
	}
	sort.SliceStable(virtualRouters, func(i, j int) bool {
		return virtualRouters[i].String() < virtualRouters[j].String()
	})
	return virtualRouters, nil
}

func (s *MeshRoutingSyncer) ReconcileVirtualNodes(c AppMeshClient, desiredVirtualNodes []appmesh.CreateVirtualNodeInput) error {
	existingMeshes, err := c.ListMeshes(nil)
	if err != nil {
		return errors.Wrapf(err, "getting existing meshes")
	}
	log.Printf("%v", existingMeshes)
	return nil
}

func (s *MeshRoutingSyncer) Try(c AppMeshClient) error {
	existingMeshes, err := c.ListMeshes(nil)
	if err != nil {
		return errors.Wrapf(err, "getting existing meshes")
	}
	log.Printf("%v", existingMeshes)
	return nil
}
