package kube

import (
	"context"
	"istio.io/istio/pkg/test/framework/components/cluster"
	"time"

	kubeApiCore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pkg/test/scopes"
	"istio.io/istio/pkg/test/util/retry"
)

var (
	defaultRetryTimeout = retry.Timeout(time.Minute * 5)
	defaultRetryDelay   = retry.Delay(time.Second * 10)
)

// GetSecret fetches a secret from the cluster
func GetSecret(cluster cluster.Cluster, secretName string, namespace string, opts ...retry.Option) (*kubeApiCore.Secret, error) {
	var secret *kubeApiCore.Secret
	_, err := retry.Do(func() (interface{}, bool, error) {
		scopes.Framework.Infof("Checking for secret %s/%s/%s...", cluster.Name(), namespace, secretName)
		fetched, err := cluster.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
		if err != nil {
			return nil, false, err
		}
		secret = fetched
		return nil, true, nil
	}, newRetryOptions(opts...)...)

	return secret, err
}

// CreateSecret creates secret in the given cluster
func CreateSecret(cluster cluster.Cluster, secret *kubeApiCore.Secret) (*kubeApiCore.Secret, error) {

	scopes.Framework.Infof("Creating secret %s/%s/%s...", cluster.Name(), secret.Namespace, secret.Name)
	s, err := cluster.CoreV1().Secrets(secret.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	return s, err
}

// DeleteSecret deletes secret in cluster
func DeleteSecret(cluster cluster.Cluster, secretName string, namespace string, opts ...retry.Option) error {
	var secret *kubeApiCore.Secret
	scopes.Framework.Infof("Deleting secret %s/%s/%s...", cluster.Name(), namespace, secretName)
	_, err := cluster.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	return err
}

func newRetryOptions(opts ...retry.Option) []retry.Option {
	out := make([]retry.Option, 0, 2+len(opts))
	out = append(out, defaultRetryTimeout, defaultRetryDelay)
	out = append(out, opts...)
	return out
}
