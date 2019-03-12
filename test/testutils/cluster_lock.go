package testutils

import (
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

var defaultService = &coreV1.Service{
	ObjectMeta: v1.ObjectMeta{
		Name:        lockServiceName,
		Annotations: make(map[string]string),
	},
}

type testClusterLocker struct {
	clientset *kubernetes.Clientset
	namespace string
	buidldId  string
}

func NewTestClusterLocker(clientset *kubernetes.Clientset, namespace, buildId string) (*testClusterLocker, error) {
	if namespace == "" {
		namespace = defaultNamespace
	}
	_, err := clientset.CoreV1().Services(namespace).Create(defaultService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}
	return &testClusterLocker{clientset: clientset, namespace: namespace, buidldId: buildId}, nil
}

func (t *testClusterLocker) AcquireLock(id string) (bool, error) {
	service, err := t.clientset.CoreV1().Services(t.namespace).Get(lockServiceName, v1.GetOptions{})
	if err != nil {
		return false, err
	}
	if _, ok := service.Annotations[lockAnnotationKey]; ok {
		return false, err
	}
	service.Annotations[lockAnnotationKey] = t.buidldId
	if _, err = t.clientset.CoreV1().Services(t.namespace).Update(service); err != nil {
		if !errors.IsConflict(err) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (t *testClusterLocker) ReleaseLock() error {
	if _, err := t.clientset.CoreV1().Services(t.namespace).Get(lockServiceName, v1.GetOptions{}); err != nil {
		return err
	}
	if _, err := t.clientset.CoreV1().Services(t.namespace).Update(defaultService); err != nil {
		return err
	}
	return nil
}
