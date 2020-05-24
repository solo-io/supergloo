package internal_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/internal"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Install namespace existence check", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("reports an error if the namespace does not exist", func() {
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		namespaceClient.EXPECT().
			GetNamespace(ctx, client.ObjectKey{Name: container_runtime.GetWriteNamespace()}).
			Return(nil, errors.NewNotFound(controllerruntime.GroupResource{}, "test-resource"))

		check := internal.NewInstallNamespaceExistenceCheck()
		clients := healthcheck_types.Clients{
			NamespaceClient: namespaceClient,
		}
		runFailure, checkApplies := check.Run(ctx, container_runtime.GetWriteNamespace(), clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.NamespaceDoesNotExist(container_runtime.GetWriteNamespace()).Error()))
		Expect(runFailure.Hint).To(Equal(fmt.Sprintf("try running `kubectl create namespace %s`", container_runtime.GetWriteNamespace())))
	})

	It("reports a generic error if the client fails", func() {
		testErr := eris.New("test-err")
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		namespaceClient.EXPECT().
			GetNamespace(ctx, client.ObjectKey{Name: container_runtime.GetWriteNamespace()}).
			Return(nil, testErr)

		check := internal.NewInstallNamespaceExistenceCheck()
		clients := healthcheck_types.Clients{
			NamespaceClient: namespaceClient,
		}
		runFailure, checkApplies := check.Run(ctx, container_runtime.GetWriteNamespace(), clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.GenericCheckFailed(testErr).Error()))
		Expect(runFailure.Hint).To(Equal("make sure the Kubernetes API server is reachable"))
	})

	It("reports success if the namespace exists", func() {
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		namespaceClient.EXPECT().
			GetNamespace(ctx, client.ObjectKey{Name: container_runtime.GetWriteNamespace()}).
			Return(nil, nil)

		runFailure, checkApplies := internal.NewInstallNamespaceExistenceCheck().Run(ctx, container_runtime.GetWriteNamespace(), healthcheck_types.Clients{
			NamespaceClient: namespaceClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).To(BeNil())
	})
})
