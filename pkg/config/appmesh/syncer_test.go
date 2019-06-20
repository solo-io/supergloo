package appmesh_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/config/appmesh"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("Syncer", func() {
	Describe("ShouldSync", func() {
		It("syncs IFF an appmesh mesh is present in either the new or old snapshot", func() {
			type testCase struct {
				old, new   *v1.ConfigSnapshot
				shouldSync bool
			}
			var testCases = []testCase{
				{
					old:        nil,
					new:        &v1.ConfigSnapshot{},
					shouldSync: false,
				},
				{
					old:        &v1.ConfigSnapshot{},
					new:        &v1.ConfigSnapshot{},
					shouldSync: false,
				},
				{
					old:        &v1.ConfigSnapshot{},
					new:        &v1.ConfigSnapshot{Meshes: v1.MeshList{inputs.IstioMesh("asdf", nil)}},
					shouldSync: false,
				},
				{
					old:        &v1.ConfigSnapshot{Meshes: v1.MeshList{inputs.IstioMesh("asdf", nil)}},
					new:        &v1.ConfigSnapshot{},
					shouldSync: false,
				},
				{
					old:        &v1.ConfigSnapshot{Meshes: v1.MeshList{inputs.AppMeshMesh("asdf", "qwer", nil)}},
					new:        &v1.ConfigSnapshot{},
					shouldSync: true,
				},
				{
					old:        &v1.ConfigSnapshot{},
					new:        &v1.ConfigSnapshot{Meshes: v1.MeshList{inputs.AppMeshMesh("asdf", "qwer", nil)}},
					shouldSync: true,
				},
				{
					old:        &v1.ConfigSnapshot{Meshes: v1.MeshList{inputs.AppMeshMesh("asdf", "qwer", nil)}},
					new:        &v1.ConfigSnapshot{Meshes: v1.MeshList{inputs.AppMeshMesh("asdf", "qwer", nil)}},
					shouldSync: true,
				},
			}
			s := NewAppMeshConfigSyncer(nil, nil, nil).(v1.ConfigSyncDeciderWithContext)
			for _, test := range testCases {
				Expect(s.ShouldSync(context.TODO(), test.old, test.new)).To(Equal(test.shouldSync))
			}
		})
	})
})
