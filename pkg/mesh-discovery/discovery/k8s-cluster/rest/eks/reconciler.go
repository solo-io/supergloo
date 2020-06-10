package eks

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	settings_utils "github.com/solo-io/service-mesh-hub/pkg/common/aws/selection"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/settings"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/common/cluster-registration"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	compute_target_aws "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/aws"
	eks_client "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/aws/clients/eks"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReconcilerName = "EKS reconciler"
	MaxResults     = 100
)

var (
	EKSClusterDiscoveryLabel = "eks-cluster-discovery"
	FailedRegisteringCluster = func(err error, name string) error {
		return eris.Wrapf(err, "Failed to register EKS cluster %s", name)
	}
	UnauthorizedForEKSCluster = func(accessKeyID string, eksClusterName string) error {
		return eris.Errorf("AWS credentials (access_key_id: %s) are not authorized for EKS cluster %s. "+
			"See https://aws.amazon.com/premiumsupport/knowledge-center/eks-api-server-unauthorized-error for details on how to enable "+
			"access for the provided credentials.", accessKeyID, eksClusterName)
	}
)

type eksReconciler struct {
	kubeClusterClient         v1alpha1.KubernetesClusterClient
	eksClientFactory          eks_client.EksClientFactory
	eksConfigBuilderFactory   eks_client.EksConfigBuilderFactory
	clusterRegistrationClient cluster_registration.ClusterRegistrationClient
	settingsClient            settings.SettingsHelperClient
	awsSelector               settings_utils.AwsSelector
}

func NewEksDiscoveryReconciler(
	kubeClusterClient v1alpha1.KubernetesClusterClient,
	eksClientFactory eks_client.EksClientFactory,
	eksConfigBuilderFactory eks_client.EksConfigBuilderFactory,
	clusterRegistrationClient cluster_registration.ClusterRegistrationClient,
	settingsClient settings.SettingsHelperClient,
	awsSelector settings_utils.AwsSelector,
) compute_target_aws.EksDiscoveryReconciler {
	return &eksReconciler{
		kubeClusterClient:         kubeClusterClient,
		eksClientFactory:          eksClientFactory,
		eksConfigBuilderFactory:   eksConfigBuilderFactory,
		clusterRegistrationClient: clusterRegistrationClient,
		settingsClient:            settingsClient,
		awsSelector:               awsSelector,
	}
}

func (e *eksReconciler) GetName() string {
	return ReconcilerName
}

func (e *eksReconciler) Reconcile(ctx context.Context, creds *credentials.Credentials, accountID string) error {
	selectorsByRegion, err := e.fetchSelectorsByRegion(ctx, accountID)
	if err != nil {
		return err
	}
	clusterNamesOnAWS := sets.NewString()
	clusterNamesOnSMH, err := e.fetchEksClustersOnSMH(ctx)
	if err != nil {
		return err
	}
	var errors *multierror.Error
	for region, selectors := range selectorsByRegion {
		eksClient, err := e.eksClientFactory(creds, region)
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}
		clusterNamesOnAWSForRegion, smhToAwsClusterNames, err := e.fetchEksClustersOnAWS(ctx, eksClient, region, selectors)
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}
		clusterNamesOnAWS = clusterNamesOnAWS.Union(clusterNamesOnAWSForRegion)
		clustersToRegister := clusterNamesOnAWSForRegion.Difference(clusterNamesOnSMH)
		for _, clusterName := range clustersToRegister.List() {
			awsClusterName := smhToAwsClusterNames[clusterName]
			err := e.registerCluster(ctx, eksClient, awsClusterName, clusterName)
			if k8s_errs.IsUnauthorized(err) {
				credsValue, err := creds.Get()
				if err != nil {
					return err
				}
				return UnauthorizedForEKSCluster(credsValue.AccessKeyID, awsClusterName)
			} else if err != nil {
				return FailedRegisteringCluster(err, awsClusterName)
			}
		}
	}
	// TODO deregister clusters that are no longer on AWS
	// clustersToDeregister := clusterNamesOnSMH.Difference(clusterNamesOnAWS)
	return errors.ErrorOrNil()
}

func (e *eksReconciler) fetchEksClustersOnAWS(
	ctx context.Context,
	eksClient cloud.EksClient,
	region string,
	selectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
) (sets.String, map[string]string, error) {
	logger := contextutils.LoggerFrom(ctx)
	var clusterNames []string
	smhToAwsClusterNames := make(map[string]string)
	input := &eks.ListClustersInput{
		MaxResults: aws.Int64(MaxResults),
	}
	for {
		logger.Debugf("Listing EKS clusters with input %+v", input)
		listClustersOutput, err := eksClient.ListClusters(ctx, input)
		if err != nil {
			return nil, nil, err
		}
		for _, clusterNameRef := range listClustersOutput.Clusters {
			clusterName := aws.StringValue(clusterNameRef)
			eksCluster, err := eksClient.DescribeCluster(ctx, clusterName)
			if err != nil {
				return nil, nil, err
			}
			matched, err := e.awsSelector.EKSMatchedBySelectors(eksCluster, selectors)
			if err != nil {
				return nil, nil, err
			}
			if !matched {
				continue
			}
			smhClusterName := metadata.BuildEksKubernetesClusterName(clusterName, region)
			clusterNames = append(clusterNames, smhClusterName)
			smhToAwsClusterNames[smhClusterName] = clusterName
		}
		if listClustersOutput.NextToken == nil {
			break
		}
		input.NextToken = listClustersOutput.NextToken
	}
	return sets.NewString(clusterNames...), smhToAwsClusterNames, nil
}

func (e *eksReconciler) fetchEksClustersOnSMH(ctx context.Context) (sets.String, error) {
	reconcilerDiscoverySelector := map[string]string{
		kube.DISCOVERED_BY: EKSClusterDiscoveryLabel,
	}
	clusters, err := e.kubeClusterClient.ListKubernetesCluster(ctx, client.MatchingLabels(reconcilerDiscoverySelector))
	if err != nil {
		return nil, err
	}
	var clusterNames []string
	for _, cluster := range clusters.Items {
		cluster := cluster
		clusterNames = append(clusterNames, cluster.GetName())
	}
	return sets.NewString(clusterNames...), nil
}

func (e *eksReconciler) registerCluster(
	ctx context.Context,
	eksClient cloud.EksClient,
	awsClusterName string,
	smhClusterName string,
) error {
	cluster, err := eksClient.DescribeCluster(ctx, awsClusterName)
	if err != nil {
		return err
	}
	config, err := e.eksConfigBuilderFactory(eksClient).ConfigForCluster(ctx, cluster)
	if err != nil {
		return err
	}
	rawConfig, err := config.RawConfig()
	if err != nil {
		return err
	}
	return e.clusterRegistrationClient.Register(
		ctx,
		config,
		smhClusterName,
		container_runtime.GetWriteNamespace(), // TODO make this configurable
		rawConfig.CurrentContext,
		EKSClusterDiscoveryLabel,
		cluster_registration.ClusterRegisterOpts{
			Overwrite:                  false,
			UseDevCsrAgentChart:        false,
			LocalClusterDomainOverride: "",
		},
	)
}

func (e *eksReconciler) fetchSelectorsByRegion(
	ctx context.Context,
	accountID string,
) (settings_utils.AwsSelectorsByRegion, error) {
	awsSettings, err := e.settingsClient.GetAWSSettingsForAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if awsSettings == nil || awsSettings.GetEksDiscovery().GetDisabled() {
		return nil, nil
	}
	if e.awsSelector.IsDiscoverAll(awsSettings.GetEksDiscovery()) ||
		(awsSettings.GetEksDiscovery() != nil && len(awsSettings.GetEksDiscovery().GetResourceSelectors()) == 0) {
		return e.awsSelector.AwsSelectorsForAllRegions(), nil
	}
	return e.awsSelector.ResourceSelectorsByRegion(awsSettings.GetEksDiscovery().GetResourceSelectors())
}
