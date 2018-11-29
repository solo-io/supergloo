package istio

import (
	"context"
	"sync"
	"unicode/utf8"

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

type MeshRoutingSyncer struct {
	lock           sync.Mutex
	activeSessions map[uint64]*appmesh.AppMesh
}

func NewMeshRoutingSyncer() *MeshRoutingSyncer {
	return &MeshRoutingSyncer{
		lock:           sync.Mutex{},
		activeSessions: make(map[uint64]*appmesh.AppMesh),
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

func (s *MeshRoutingSyncer) NewOrCachedClient(appMesh *v1.AppMesh, secrets gloov1.SecretList) (*appmesh.AppMesh, error) {
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

	var desiredMeshes []*appmesh.CreateMeshInput
	for _, mesh := range snap.Meshes.List() {
		appMesh, ok := mesh.MeshType.(*v1.Mesh_AppMesh)
		if !ok {
			continue
		}
		if appMesh.AppMesh == nil {
			return errors.Errorf("%v missing configuration for AppMesh", mesh.Metadata.Ref())
		}

		desiredMesh := &appmesh.CreateMeshInput{
			MeshName: aws.String(mesh.Metadata.Name),
		}
		desiredMeshes = append(desiredMeshes, desiredMesh)
	}

	/*
		0 - mesh per mesh
		1 - virtual node per upstream
		2 - routing rules get aggregated into virtual service like object
		routes on virtual service become the aws routes
		virtual service becomes virtual router
	*/
	exampleMesh := appmesh.CreateMeshInput{}
	exampleVirtualNode := appmesh.CreateVirtualNodeInput{}
	exampleVirtualRouter := appmesh.CreateVirtualRouterInput{}
	exampleRoute := appmesh.CreateRouteInput{}

	destinationRules, err := virtualNodesForUpstreams(rules, meshes, upstreams)
	if err != nil {
		return errors.Wrapf(err, "creating subsets from snapshot")
	}

	virtualServices, err := virtualServicesForRules(rules, meshes, upstreams)
	if err != nil {
		return errors.Wrapf(err, "creating virtual services from snapshot")
	}
	return s.writeIstioCrds(ctx, destinationRules, virtualServices)
}
