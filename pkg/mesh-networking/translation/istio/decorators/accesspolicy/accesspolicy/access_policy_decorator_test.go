package accesspolicy_test

import (
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha22 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy"
	accesspolicy2 "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy/accesspolicy"
	securityv1beta1spec "istio.io/api/security/v1beta1"
)

var _ = Describe("AccessPolicyDecorator", func() {
	var (
		accessPolicyDecorator accesspolicy.AuthorizationPolicyDecorator
	)

	BeforeEach(func() {
		accessPolicyDecorator = accesspolicy2.NewAccessPolicyDecorator()
	})

	It("should set paths, methods, and ports", func() {
		registerFieldFunc := func(fieldPtr, val interface{}) error {
			return nil
		}
		output := &securityv1beta1spec.Operation{}
		appliedPolicy := &v1alpha2.MeshServiceStatus_AppliedAccessPolicy{
			Spec: &v1alpha22.AccessPolicySpec{
				AllowedPaths: []string{
					"path1", "path2",
				},
				AllowedMethods: []types.HttpMethodValue{
					types.HttpMethodValue_GET, types.HttpMethodValue_DELETE,
				},
				AllowedPorts: []uint32{
					9080, 8080,
				},
			},
		}
		err := accessPolicyDecorator.ApplyToAuthorizationPolicy(
			appliedPolicy,
			nil,
			output,
			registerFieldFunc,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(Equal(&securityv1beta1spec.Operation{
			Paths: appliedPolicy.Spec.AllowedPaths,
			Methods: []string{
				types.HttpMethodValue_GET.String(), types.HttpMethodValue_DELETE.String(),
			},
			Ports: []string{
				"9080", "8080",
			},
		}))
	})

	It("should set allowed methods to '*' if no allowedMethods on AccessPolicy", func() {
		registerFieldFunc := func(fieldPtr, val interface{}) error {
			return nil
		}
		output := &securityv1beta1spec.Operation{}
		appliedPolicy := &v1alpha2.MeshServiceStatus_AppliedAccessPolicy{
			Spec: &v1alpha22.AccessPolicySpec{},
		}
		err := accessPolicyDecorator.ApplyToAuthorizationPolicy(
			appliedPolicy,
			nil,
			output,
			registerFieldFunc,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(Equal(&securityv1beta1spec.Operation{
			Methods: []string{"*"},
		}))
	})

	It("should not set any fields if error encountered while registering field", func() {
		testErr := eris.New("register field error")
		registerFieldFunc := func(fieldPtr, val interface{}) error {
			return testErr
		}
		output := &securityv1beta1spec.Operation{}
		appliedPolicy := &v1alpha2.MeshServiceStatus_AppliedAccessPolicy{
			Spec: &v1alpha22.AccessPolicySpec{},
		}
		err := accessPolicyDecorator.ApplyToAuthorizationPolicy(
			appliedPolicy,
			nil,
			output,
			registerFieldFunc,
		)
		// One error for each field registration attempt.
		Expect(err).To(Equal(multierror.Append(&multierror.Error{}, []error{testErr, testErr, testErr}...)))
		Expect(output).To(Equal(&securityv1beta1spec.Operation{}))
	})
})
