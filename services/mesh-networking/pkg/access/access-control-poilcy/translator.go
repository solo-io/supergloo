package access_control_poilcy

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	"github.com/solo-io/mesh-projects/pkg/logging"
	securityv1beta1 "istio.io/api/security/v1beta1"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewAccessControlPolicyTranslator(
	controller controller.AccessControlPolicyController,
	authPolicyClient security.AuthorizationPolicyClient,
) AccessControlPolicyTranslator {
	return &translator{
		controller:       controller,
		authPolicyClient: authPolicyClient,
	}
}

type translator struct {
	controller       controller.AccessControlPolicyController
	authPolicyClient security.AuthorizationPolicyClient
}

func (t *translator) Start(ctx context.Context) error {
	return t.controller.AddEventHandler(ctx, &controller.AccessControlPolicyEventHandlerFuncs{
		OnCreate: func(policy *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, policy)
			logger.Debugf("Handling event: %+v", policy)

			err := t.authPolicyClient.Upsert(ctx, &v1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      getResourceName(policy.ObjectMeta),
					Namespace: policy.GetNamespace(),
				},
				Spec: securityv1beta1.AuthorizationPolicy{},
			})
			if err != nil {
				logger.Error(err)
			}
			return nil
		},
		OnUpdate: func(_, policy *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, policy)
			logger.Debugf("Handling event: %+v", policy)

			err := t.authPolicyClient.Upsert(ctx, &v1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      getResourceName(policy.ObjectMeta),
					Namespace: policy.GetNamespace(),
				},
				Spec: securityv1beta1.AuthorizationPolicy{},
			})
			if err != nil {
				logger.Error("Error while handling AccessControlPolicy update event", err)
			}
			return nil
		},
		OnDelete: func(policy *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, policy)
			logger.Warn("Ignoring event: %+v", policy)
			return nil
		},
		OnGeneric: func(policy *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.GenericEvent, policy)
			logger.Warn("Ignoring event: %+v", policy)
			return nil
		},
	})
}

func getResourceName(meta v1.ObjectMeta) string {
	return meta.Name + "-" + meta.ClusterName
}
