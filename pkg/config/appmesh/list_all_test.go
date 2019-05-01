package appmesh_test

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/errors"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mocks "github.com/solo-io/supergloo/pkg/config/appmesh/mocks"

	. "github.com/solo-io/supergloo/pkg/config/appmesh"
)

var _ = Describe("List all App Mesh resources", func() {

	var (
		ctrl *gomock.Controller
		mesh = "test-mesh"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("correctly lists all resources for a mesh", func() {
		mockClient := mocks.NewMockClient(ctrl)
		mockClient.EXPECT().ListVirtualNodes(gomock.Any(), mesh).DoAndReturn(returnWithDelay("vn-1", "vn-2", "vn-3")).Times(1)
		mockClient.EXPECT().ListVirtualServices(gomock.Any(), mesh).DoAndReturn(returnWithDelay("vs-1", "vs-2", "vs-3")).Times(1)
		mockClient.EXPECT().ListVirtualRouters(gomock.Any(), mesh).DoAndReturn(returnWithDelay("vr-1", "vr-2")).Times(1)
		mockClient.EXPECT().ListRoutes(gomock.Any(), mesh, "vr-1").DoAndReturn(returnRoutesWithDelay("vr1-route-1")).Times(1)
		mockClient.EXPECT().ListRoutes(gomock.Any(), mesh, "vr-2").DoAndReturn(returnRoutesWithDelay("vr2-route-1", "vr2-route-2")).Times(1)

		resources, err := ListAllForMesh(context.TODO(), mockClient, mesh)
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).NotTo(BeNil())

		Expect(resources.VirtualNodes).To(ConsistOf("vn-1", "vn-2", "vn-3"))
		Expect(resources.VirtualServices).To(ConsistOf("vs-1", "vs-2", "vs-3"))

		Expect(resources.VirtualRouters).To(HaveLen(2))

		routesVr1, ok := resources.VirtualRouters["vr-1"]
		Expect(ok).To(BeTrue())
		Expect(routesVr1).To(ConsistOf("vr1-route-1"))

		routesVr2, ok := resources.VirtualRouters["vr-2"]
		Expect(ok).To(BeTrue())
		Expect(routesVr2).To(ConsistOf("vr2-route-1", "vr2-route-2"))
	})

	It("fails if an error occurs in any call to the underlying client", func() {
		mockClient := mocks.NewMockClient(ctrl)
		mockClient.EXPECT().ListVirtualNodes(gomock.Any(), mesh).DoAndReturn(returnWithDelay("vn-1", "vn-2", "vn-3")).Times(1)
		mockClient.EXPECT().ListVirtualServices(gomock.Any(), mesh).DoAndReturn(failWithDelay()).Times(1)
		mockClient.EXPECT().ListVirtualRouters(gomock.Any(), mesh).DoAndReturn(returnWithDelay("vr-1", "vr-2")).Times(1)
		mockClient.EXPECT().ListRoutes(gomock.Any(), mesh, "vr-1").DoAndReturn(returnRoutesWithDelay("vr1-route-1")).Times(1)
		mockClient.EXPECT().ListRoutes(gomock.Any(), mesh, "vr-2").DoAndReturn(returnRoutesWithDelay("vr2-route-1", "vr2-route-2")).Times(1)

		_, err := ListAllForMesh(context.TODO(), mockClient, mesh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to list all App Mesh resources for mesh test-mesh"))
		Expect(err.Error()).To(ContainSubstring("1 error occurred"))
		Expect(err.Error()).To(ContainSubstring("simulating failure"))
	})

	It("fails and logs correctly if multiple errors occurs in calls to the underlying client", func() {
		mockClient := mocks.NewMockClient(ctrl)
		mockClient.EXPECT().ListVirtualNodes(gomock.Any(), mesh).DoAndReturn(returnWithDelay("vn-1", "vn-2", "vn-3")).Times(1)
		mockClient.EXPECT().ListVirtualServices(gomock.Any(), mesh).DoAndReturn(failWithDelay()).Times(1)
		mockClient.EXPECT().ListVirtualRouters(gomock.Any(), mesh).DoAndReturn(returnWithDelay("vr-1", "vr-2")).Times(1)
		mockClient.EXPECT().ListRoutes(gomock.Any(), mesh, "vr-1").DoAndReturn(failRoutesWithDelay()).Times(1)
		mockClient.EXPECT().ListRoutes(gomock.Any(), mesh, "vr-2").DoAndReturn(returnRoutesWithDelay("vr2-route-1", "vr2-route-2")).Times(1)

		_, err := ListAllForMesh(context.TODO(), mockClient, mesh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to list all App Mesh resources for mesh test-mesh"))
		Expect(err.Error()).To(ContainSubstring("2 errors occurred"))
		Expect(err.Error()).To(ContainSubstring("simulating failure"))
		Expect(err.Error()).To(ContainSubstring("simulating failure on route"))
	})
})

// Represents the signature of ListVirtualNode/ListVirtualServices/ListVirtualRouters functions
type listVnVsVrFunc func(ctx context.Context, mesh string) ([]string, error)

// Represents the signature of ListRoutes function
type listRoutesFunc func(ctx context.Context, mesh, vr string) ([]string, error)

var returnWithDelay = func(names ...string) listVnVsVrFunc {
	return func(context.Context, string) ([]string, error) {
		time.Sleep(time.Duration(rand.IntnRange(50, 500)) * time.Millisecond)
		return names, nil
	}
}
var returnRoutesWithDelay = func(routeNames ...string) listRoutesFunc {
	return func(context.Context, string, string) ([]string, error) {
		time.Sleep(time.Duration(rand.IntnRange(50, 500)) * time.Millisecond)
		return routeNames, nil
	}
}

var failWithDelay = func() listVnVsVrFunc {
	return func(context.Context, string) ([]string, error) {
		time.Sleep(time.Duration(rand.IntnRange(50, 500)) * time.Millisecond)
		return nil, errors.Errorf("simulating failure")
	}
}
var failRoutesWithDelay = func() listRoutesFunc {
	return func(context.Context, string, string) ([]string, error) {
		time.Sleep(time.Duration(rand.IntnRange(50, 500)) * time.Millisecond)
		return nil, errors.Errorf("simulating failure on route")
	}
}
