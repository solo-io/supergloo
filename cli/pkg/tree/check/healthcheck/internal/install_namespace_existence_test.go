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
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
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
			Get(ctx, env.DefaultWriteNamespace).
			Return(nil, errors.NewNotFound(controllerruntime.GroupResource{}, "test-resource"))

		check := internal.NewInstallNamespaceExistenceCheck()
		clients := healthcheck_types.Clients{
			NamespaceClient: namespaceClient,
		}
		runFailure, checkApplies := check.Run(ctx, env.DefaultWriteNamespace, clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.NamespaceDoesNotExist(env.DefaultWriteNamespace).Error()))
		Expect(runFailure.Hint).To(Equal(fmt.Sprintf("try running `kubectl create namespace %s`", env.DefaultWriteNamespace)))
	})

	It("reports a generic error if the client fails", func() {
		testErr := eris.New("test-err")
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		namespaceClient.EXPECT().
			Get(ctx, env.DefaultWriteNamespace).
			Return(nil, testErr)

		check := internal.NewInstallNamespaceExistenceCheck()
		clients := healthcheck_types.Clients{
			NamespaceClient: namespaceClient,
		}
		runFailure, checkApplies := check.Run(ctx, env.DefaultWriteNamespace, clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.GenericCheckFailed(testErr).Error()))
		Expect(runFailure.Hint).To(Equal("make sure the Kubernetes API server is reachable"))
	})

	It("reports success if the namespace exists", func() {
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		namespaceClient.EXPECT().
			Get(ctx, env.DefaultWriteNamespace).
			Return(nil, nil)

		runFailure, checkApplies := internal.NewInstallNamespaceExistenceCheck().Run(ctx, env.DefaultWriteNamespace, healthcheck_types.Clients{
			NamespaceClient: namespaceClient,
		})

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).To(BeNil())
	})
})
