package istio

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/go-multierror"

	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	glookubev1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/kubernetes"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type PolicySyncer struct {
	WriteSelector  map[string]string // for reconciling only our resources
	WriteNamespace string

	serviceRoleBindingReconciler v1alpha1.ServiceRoleBindingReconciler
	serviceRoleReconciler        v1alpha1.ServiceRoleReconciler
	rbacConfigReconciler         v1alpha1.RbacConfigReconciler

	kubeClient *kubernetes.Clientset
}

func NewPolicySyncer(writens string, kubeCache *kube.KubeCache, restConfig *rest.Config) (*PolicySyncer, error) {
	var ps PolicySyncer
	ps.WriteNamespace = writens

	serviceRoleBindingClient, err := v1alpha1.NewServiceRoleBindingClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha1.ServiceRoleBindingCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := serviceRoleBindingClient.Register(); err != nil {
		return nil, err
	}
	ps.serviceRoleBindingReconciler = v1alpha1.NewServiceRoleBindingReconciler(serviceRoleBindingClient)

	serviceRoleClient, err := v1alpha1.NewServiceRoleClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha1.ServiceRoleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := serviceRoleClient.Register(); err != nil {
		return nil, err
	}
	ps.serviceRoleReconciler = v1alpha1.NewServiceRoleReconciler(serviceRoleClient)

	rbacConfigClient, err := v1alpha1.NewRbacConfigClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha1.RbacConfigCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := rbacConfigClient.Register(); err != nil {
		return nil, err
	}
	ps.rbacConfigReconciler = v1alpha1.NewRbacConfigReconciler(rbacConfigClient)

	ps.kubeClient, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &ps, nil

}

func (s *PolicySyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	var multiErr *multierror.Error

	for _, mesh := range snap.Meshes.List() {
		_, ok := mesh.MeshType.(*v1.Mesh_Istio)
		if !ok {
			// not our mesh, we don't care
			continue
		}
		policy := mesh.Policy
		if policy == nil {
			err := s.removePolicy(ctx)
			if err != nil {
				multiErr = multierror.Append(multiErr, err)
			}
			continue
		}

		err := s.syncPolicy(ctx, snap.Upstreams, policy)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}

func (s *PolicySyncer) removePolicy(ctx context.Context) error {
	opts := clients.ListOpts{
		Ctx:      ctx,
		Selector: s.WriteSelector,
	}

	// delete everything!
	err := s.rbacConfigReconciler.Reconcile(s.WriteNamespace, nil, preserveRbacConfig, opts)
	if err != nil {
		return err
	}

	// get all namespaces
	namespaces, err := s.kubeClient.CoreV1().Namespaces().List(kubemeta.ListOptions{})
	if err != nil {
		return err
	}

	for _, namespace := range namespaces.Items {
		ns := namespace.Name
		err := s.serviceRoleBindingReconciler.Reconcile(ns, nil, preserveServiceRoleBinding, opts)
		if err != nil {
			return err
		}

		err = s.serviceRoleReconciler.Reconcile(ns, nil, preserveServiceRole, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PolicySyncer) syncPolicy(ctx context.Context, upstreams gloov1.UpstreamsByNamespace, p *v1.Policy) error {
	// go over all the available namespaces and reconcile.

	opts := clients.ListOpts{
		Ctx:      ctx,
		Selector: s.WriteSelector,
	}

	// we have a policy, write a global config
	rcfg := s.globalConfig()
	var rcfgs v1alpha1.RbacConfigList
	rcfgs = append(rcfgs, rcfg)
	converter := convertToIstio{upstreams, p, s.kubeClient}
	serviceRoles, serviceRolesBindings := converter.toIstio()

	resources.UpdateMetadata(rcfg, s.updateMetadata)
	for _, res := range serviceRoles {
		resources.UpdateMetadata(res, s.updateMetadata)
	}
	for _, res := range serviceRolesBindings {
		resources.UpdateMetadata(res, s.updateMetadata)
	}

	// get all namespaces
	namespaces, err := s.kubeClient.CoreV1().Namespaces().List(kubemeta.ListOptions{})
	if err != nil {
		return err
	}

	for _, namespace := range namespaces.Items {
		currentns := namespace.Name
		var currentsrb []*v1alpha1.ServiceRoleBinding
		for _, srb := range serviceRolesBindings {
			if srb.Metadata.Namespace == currentns {
				currentsrb = append(currentsrb, srb)
			}
		}
		var currentsr []*v1alpha1.ServiceRole
		for _, sr := range serviceRoles {
			if sr.Metadata.Namespace == currentns {
				currentsr = append(currentsr, sr)
			}
		}

		err = s.serviceRoleBindingReconciler.Reconcile(currentns, currentsrb, preserveServiceRoleBinding, opts)
		if err != nil {
			return err
		}
		err = s.serviceRoleReconciler.Reconcile(currentns, currentsr, preserveServiceRole, opts)
		if err != nil {
			return err
		}
	}

	err = s.rbacConfigReconciler.Reconcile(s.WriteNamespace, rcfgs, preserveRbacConfig, opts)
	if err != nil {
		return err
	}
	return nil

}

func (s *PolicySyncer) globalConfig() *v1alpha1.RbacConfig {
	return &v1alpha1.RbacConfig{
		Metadata: core.Metadata{
			// name MUST be default.
			Name:      "default",
			Namespace: s.WriteNamespace,
		},
		Mode:            v1alpha1.RbacConfig_ON,
		EnforcementMode: v1alpha1.EnforcementMode_ENFORCED,
	}
}

type convertToIstio struct {
	upstreams  gloov1.UpstreamsByNamespace
	policy     *v1.Policy
	kubeClient *kubernetes.Clientset
}

func (c *convertToIstio) toIstio() ([]*v1alpha1.ServiceRole, []*v1alpha1.ServiceRoleBinding) {
	var roles []*v1alpha1.ServiceRole
	var bindings []*v1alpha1.ServiceRoleBinding

	rulesByDest := map[core.ResourceRef][]*v1.Rule{}
	for _, rule := range c.policy.Rules {
		if rule.Source == nil {
			// TODO: should we return error instead?
			continue
		}
		if rule.Destination == nil {
			// TODO: should we return error instead?
			continue
		}
		rulesByDest[*rule.Destination] = append(rulesByDest[*rule.Destination], rule)
	}
	// sort for idempotency
	for _, rule := range c.policy.Rules {
		dests := rulesByDest[*rule.Destination]
		sort.Slice(dests, func(i, j int) bool {
			return dests[i].Source.String() > dests[j].Source.String()
		})
	}

	for dest, rules := range rulesByDest {

		destupstream := c.getkube(dest)
		if destupstream == nil {
			continue
		}
		var destref core.ResourceRef
		destref.Name = destupstream.ServiceName
		destref.Namespace = destupstream.ServiceNamespace

		// objects need to be written to the same namespaces as the service the control.
		ns := destupstream.ServiceNamespace
		// create an istio service role and binding:
		name := "access-" + dest.Namespace + "-" + dest.Name
		// create service role:
		sr := &v1alpha1.ServiceRole{
			Metadata: core.Metadata{
				Name:      name,
				Namespace: ns,
			},
			Rules: []*v1alpha1.AccessRule{
				{
					Methods: []string{"*"},
					Services: []string{
						c.svcname(destref),
					},
				},
			},
		}
		var subjects []*v1alpha1.Subject
		for _, rule := range rules {
			sourceupstream := c.getkube(*rule.Source)
			if sourceupstream == nil {
				continue
			}

			sa := c.getsvcaccount(sourceupstream)
			if sa == nil {
				continue
			}

			subjects = append(subjects, &v1alpha1.Subject{
				Properties: map[string]string{
					"source.principal": c.principalame(*sa),
				},
			})
		}
		name = "bind-" + dest.Namespace + "-" + dest.Name
		srb := &v1alpha1.ServiceRoleBinding{
			Metadata: core.Metadata{
				Name:      name,
				Namespace: ns,
			},

			Subjects: subjects,
			RoleRef: &v1alpha1.RoleRef{
				Kind: "ServiceRole",
				Name: sr.Metadata.Name,
			},
		}
		roles = append(roles, sr)
		bindings = append(bindings, srb)
	}
	return roles, bindings
}

func (c *convertToIstio) getsvcaccount(k *glookubev1.UpstreamSpec) *core.ResourceRef {
	// istio manages identity in the level of service accounts.
	// so we hueristicly figure out the service account for this upstream.
	// we may consider changing our API in the future to better support this usecase

	svcname := k.ServiceName
	svcnamespace := k.ServiceNamespace

	// find the services and get the selectors, and
	svc, err := c.kubeClient.CoreV1().Services(svcnamespace).Get(svcname, kubemeta.GetOptions{})
	if err != nil {
		return nil
	}
	// get the pods from the selector
	// get the first pod and grab its service account
	opts := kubemeta.ListOptions{
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
		Limit:         1,
	}
	pods, err := c.kubeClient.CoreV1().Pods(svcnamespace).List(opts)
	if err != nil {
		return nil
	}

	if len(pods.Items) == 0 {
		return nil
	}
	saname := pods.Items[0].Spec.ServiceAccountName
	return &core.ResourceRef{
		Name:      saname,
		Namespace: svcnamespace,
	}
}

func (c *convertToIstio) getkube(ref core.ResourceRef) *glookubev1.UpstreamSpec {
	upstream, err := c.upstreams.List().Find(ref.Namespace, ref.Name)
	if err != nil {
		return nil
	}
	kubeupstream, ok := upstream.UpstreamSpec.UpstreamType.(*gloov1.UpstreamSpec_Kube)
	if !ok {
		return nil
	}
	return kubeupstream.Kube
}

func (c *convertToIstio) svcname(s core.ResourceRef) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", s.Name, s.Namespace)
}
func (c *convertToIstio) principalame(s core.ResourceRef) string {
	return fmt.Sprintf("cluster.local/ns/%s/sa/%s", s.Namespace, s.Name)
}

func (s *PolicySyncer) updateMetadata(meta *core.Metadata) {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if meta.Labels == nil && len(s.WriteSelector) > 0 {
		meta.Labels = make(map[string]string)
	}
	meta.Annotations["created_by"] = "supergloo"
	for k, v := range s.WriteSelector {
		meta.Labels[k] = v
	}
}

func preserveServiceRoleBinding(original, desired *v1alpha1.ServiceRoleBinding) (bool, error) {
	original.Metadata = desired.Metadata
	original.Status = desired.Status
	return !proto.Equal(original, desired), nil
}
func preserveServiceRole(original, desired *v1alpha1.ServiceRole) (bool, error) {
	original.Metadata = desired.Metadata
	original.Status = desired.Status
	return !proto.Equal(original, desired), nil
}
func preserveRbacConfig(original, desired *v1alpha1.RbacConfig) (bool, error) {
	original.Metadata = desired.Metadata
	original.Status = desired.Status
	return !proto.Equal(original, desired), nil
}
