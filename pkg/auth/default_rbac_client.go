package auth

import (
	"fmt"

	k8sapiv1 "k8s.io/api/core/v1"
	rbactypes "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8srbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

type defaultRbacClient struct {
	clusterRoleBindingClient k8srbacv1.ClusterRoleBindingInterface
}

func (r *defaultRbacClient) BindClusterRolesToServiceAccount(targetServiceAccount *k8sapiv1.ServiceAccount, roles []*rbactypes.ClusterRole) error {
	for _, role := range roles {
		_, err := r.clusterRoleBindingClient.Create(&rbactypes.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s-clusterrole-binding", targetServiceAccount.GetName(), role.GetName()),
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			Subjects: []rbactypes.Subject{{
				Kind:      "ServiceAccount",
				Name:      targetServiceAccount.GetName(),
				Namespace: targetServiceAccount.GetNamespace(),
			}},
			RoleRef: rbactypes.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     role.GetName(),
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}
