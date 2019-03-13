package testutils

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// default namespace to install
	defaultNamespace = "default"
	// name of kubermnetes service holding the lock
	lockServiceName = "test-lock"
	// name of the annotation containing the lock
	lockAnnotationKey = "test.lock"
)

var defaultConfigMap = &coreV1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name: lockServiceName,
	},
}

var defaultOpts = []retry.Option{
	retry.Delay(30 * time.Second),
	retry.Attempts(20),
	retry.DelayType(retry.FixedDelay),
	retry.RetryIf(func(e error) bool {
		if e != nil {
			if IsLockIsUseError(e) {
				return true
			}
		}
		return false
	}),
}

var lockInUseError = fmt.Errorf("lock is current in use")
var IsLockIsUseError = func(e error) bool {
	return e == lockInUseError
}

type testClusterLocker struct {
	clientset kubernetes.Interface
	namespace string
	buidldId  string
}

func NewTestClusterLocker(clientset kubernetes.Interface, namespace, buildId string) (*testClusterLocker, error) {
	if namespace == "" {
		namespace = defaultNamespace
	}
	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(defaultConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}
	return &testClusterLocker{clientset: clientset, namespace: namespace, buidldId: buildId}, nil
}

func (t *testClusterLocker) AcquireLock(opts ...retry.Option) error {
	opts = append(defaultOpts, opts...)
	err := retry.Do(
		func() error {
			cfgMap, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Get(lockServiceName, v1.GetOptions{})
			if err != nil && !errors.IsTimeout(err) {
				return err
			}

			if cfgMap.Annotations == nil || len(cfgMap.Annotations) == 0 {
				cfgMap.Annotations = map[string]string{
					lockAnnotationKey: t.buidldId,
				}
			} else {
				if val, ok := cfgMap.Annotations[lockAnnotationKey]; ok && val != t.buidldId {
					return lockInUseError
				}
			}

			if _, err = t.clientset.CoreV1().ConfigMaps(t.namespace).Update(cfgMap); err != nil {
				if !errors.IsConflict(err) {
					return err
				}
			}
			return nil
		},
		opts...,
	)

	return err

}

func (t *testClusterLocker) ReleaseLock() error {
	if _, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Get(lockServiceName, v1.GetOptions{}); err != nil {
		return err
	}
	if _, err := t.clientset.CoreV1().ConfigMaps(t.namespace).Update(defaultConfigMap); err != nil {
		return err
	}
	return nil
}
