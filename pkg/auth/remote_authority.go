package auth

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	k8sapiv1 "k8s.io/api/core/v1"
	rbactypes "k8s.io/api/rbac/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRemoteAuthorityManager(
	serviceAccountClient kubernetes_core.ServiceAccountClient,
	rbacClient RbacClient,
) RemoteAuthorityManager {

	return &remoteAuthorityManager{
		serviceAccountClient: serviceAccountClient,
		rbacClient:           rbacClient,
	}
}

type remoteAuthorityManager struct {
	serviceAccountClient kubernetes_core.ServiceAccountClient
	rbacClient           RbacClient
}

func (r *remoteAuthorityManager) ApplyRemoteServiceAccount(
	ctx context.Context,
	newServiceAccountRef *types.ResourceRef,
	roles []*rbactypes.ClusterRole,
) (*k8sapiv1.ServiceAccount, error) {

	saToCreate := &k8sapiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newServiceAccountRef.GetName(),
			Namespace: newServiceAccountRef.GetNamespace(),
		},
	}

	err := r.serviceAccountClient.Create(ctx, saToCreate)
	if err != nil {
		if kubeerrs.IsAlreadyExists(err) {
			err = r.serviceAccountClient.Update(ctx, saToCreate)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	err = r.rbacClient.BindClusterRolesToServiceAccount(saToCreate, roles)
	if err != nil {
		return nil, err
	}

	return saToCreate, nil
}
