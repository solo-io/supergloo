package auth

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/constants"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_rbac_types "k8s.io/api/rbac/v1"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RegistrationServiceAccount      = "solo.io/registration-service-account"
	RegistrationServiceAccountValue = "true"
)

func NewRemoteAuthorityManager(
	serviceAccountClient k8s_core.ServiceAccountClient,
	rbacClient RbacClient,
) RemoteAuthorityManager {

	return &remoteAuthorityManager{
		serviceAccountClient: serviceAccountClient,
		rbacClient:           rbacClient,
	}
}

type remoteAuthorityManager struct {
	serviceAccountClient k8s_core.ServiceAccountClient
	rbacClient           RbacClient
}

func (r *remoteAuthorityManager) ApplyRemoteServiceAccount(
	ctx context.Context,
	newServiceAccountRef *zephyr_core_types.ResourceRef,
	roles []*k8s_rbac_types.ClusterRole,
) (*k8s_core_types.ServiceAccount, error) {

	saToCreate := &k8s_core_types.ServiceAccount{
		ObjectMeta: k8s_meta.ObjectMeta{
			Name:      newServiceAccountRef.GetName(),
			Namespace: newServiceAccountRef.GetNamespace(),
			Labels: map[string]string{
				constants.ManagedByLabel:   constants.ServiceMeshHubApplicationName,
				RegistrationServiceAccount: RegistrationServiceAccountValue,
			},
		},
	}

	err := r.serviceAccountClient.CreateServiceAccount(ctx, saToCreate)
	if err != nil {
		if k8s_errs.IsAlreadyExists(err) {
			err = r.serviceAccountClient.UpdateServiceAccount(ctx, saToCreate)
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
