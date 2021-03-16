package utils

import (
	"context"

	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureNamespace(ctx context.Context, kubeClient client.Client, namespace string) error {
	namespaces := v1.NewNamespaceClient(kubeClient)
	return namespaces.UpsertNamespace(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{Finalizers: []corev1.FinalizerName{"kubernetes"}},
	})
}
