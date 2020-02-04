package upgrade_assets_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets"
	mock_upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets/mocks"
)

var _ = Describe("upgrade assets helper", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})
	Context("asset helper", func() {
		var (
			helper upgrade_assets.AssetHelper

			client *mock_upgrade_assets.MockGithubAssetClient
		)

		BeforeEach(func() {
			client = mock_upgrade_assets.NewMockGithubAssetClient(ctrl)
			helper = upgrade_assets.NewAssetHelper(client)
		})

		Context("download asset", func() {
			It("will return an error with an illegal url", func() {
				downloadUrl := "illegal-url"
				err := helper.DownloadAsset(downloadUrl, "")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("get release with asset", func() {
			It("will return an error if a tag is supplied and no release is found", func() {
				tag := "vtest"
				client.EXPECT().GetReleaseByTag(ctx, tag).Return(nil, testErr)
				_, err := helper.GetReleaseWithAsset(ctx, tag, upgrade_assets.MeshctlBinaryName())
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})

			It("will return an error if list releases fails", func() {
				tag := "latest"
				client.EXPECT().ListReleases(ctx, gomock.Any()).Return(nil, testErr)
				_, err := helper.GetReleaseWithAsset(ctx, tag, upgrade_assets.MeshctlBinaryName())
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(testErr))
			})

			It("will return a release if it can be found in the releases", func() {
				tag := "latest"
				targetRelease := &github.RepositoryRelease{
					Assets: []github.ReleaseAsset{
						{
							Name: github.String(upgrade_assets.MeshctlBinaryName()),
						},
					},
				}
				client.EXPECT().ListReleases(ctx, gomock.Any()).Return([]*github.RepositoryRelease{
					targetRelease,
				}, nil)
				release, err := helper.GetReleaseWithAsset(ctx, tag, upgrade_assets.MeshctlBinaryName())
				Expect(err).NotTo(HaveOccurred())
				Expect(release).To(Equal(targetRelease))
			})
		})

	})
})
