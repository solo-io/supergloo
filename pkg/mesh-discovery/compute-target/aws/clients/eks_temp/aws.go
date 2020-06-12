package eks_temp

import (
	"context"
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

type EksClient interface {
	DescribeCluster(ctx context.Context, name string) (*eks.Cluster, error)
	ListClusters(ctx context.Context, input *eks.ListClustersInput) (*eks.ListClustersOutput, error)
	Token(ctx context.Context, name string) (token.Token, error)
}

type EksConfigBuilder interface {
	ConfigForCluster(ctx context.Context, cluster *eks.Cluster) (clientcmd.ClientConfig, error)
}

func NewEksClient(sess *session.Session) EksClient {
	return &awsClient{
		sess: sess,
	}
}

type awsClient struct {
	sess *session.Session
}

func (a *awsClient) Session() *session.Session {
	return a.sess
}

func (a *awsClient) DescribeCluster(ctx context.Context, name string) (*eks.Cluster, error) {
	eksSvc := eks.New(a.sess)
	input := &eks.DescribeClusterInput{
		Name: aws.String(name),
	}
	resp, err := eksSvc.DescribeClusterWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.Cluster, nil
}

func (a *awsClient) ListClusters(ctx context.Context, input *eks.ListClustersInput) (*eks.ListClustersOutput, error) {
	eksSvc := eks.New(a.sess)
	return eksSvc.ListClustersWithContext(ctx, input)
}

func (a *awsClient) Token(ctx context.Context, name string) (token.Token, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return token.Token{}, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: name,
		Session:   a.sess,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return token.Token{}, err
	}
	return tok, nil
}

func NewEksConfigBuilder(eksClient cloud.EksClient) EksConfigBuilder {
	return &awsClusterConfigBuilder{
		eksClient: eksClient,
	}
}

type awsClusterConfigBuilder struct {
	eksClient cloud.EksClient
}

func (a *awsClusterConfigBuilder) ConfigForCluster(ctx context.Context, cluster *eks.Cluster) (clientcmd.ClientConfig, error) {
	tok, err := a.eksClient.Token(ctx, aws.StringValue(cluster.Name))
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}

	cfg := kubeconfig.BuildRemoteCfg(
		&clientcmdapi.Cluster{
			Server:                   aws.StringValue(cluster.Endpoint),
			CertificateAuthorityData: ca,
		},
		&clientcmdapi.Context{
			Cluster: aws.StringValue(cluster.Name),
		},
		aws.StringValue(cluster.Name),
		tok.Token,
	)

	return clientcmd.NewDefaultClientConfig(cfg, &clientcmd.ConfigOverrides{}), nil
}
