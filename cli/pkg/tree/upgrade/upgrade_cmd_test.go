package upgrade_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade"
	upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets"
	mock_upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets/mocks"
)

var _ = Describe("Upgrade", func() {
	var (
		ctrl         *gomock.Controller
		ctx          context.Context
		meshctl      *cli_mocks.MockMeshctl
		mockUpgrader *mock_upgrade_assets.MockAssetHelper

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockUpgrader = mock_upgrade_assets.NewMockAssetHelper(ctrl)
		meshctl = &cli_mocks.MockMeshctl{
			MockController: ctrl,
			Clients: common.Clients{
				ReleaseAssetHelper: mockUpgrader,
			},
			Ctx: ctx,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will error if release cannot be fetched", func() {
		opts := options.Upgrade{
			ReleaseTag: cliconstants.DefaultReleaseTag,
		}
		mockUpgrader.EXPECT().GetReleaseWithAsset(meshctl.Ctx, opts.ReleaseTag, upgrade_assets.MeshctlBinaryName()).Return(nil, testErr)
		_, err := meshctl.Invoke("upgrade")
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(upgrade.GetReleaseForRepoError(testErr, opts.ReleaseTag)))
	})

	It("will error if release the asset cannot be found", func() {
		opts := options.Upgrade{
			ReleaseTag: cliconstants.DefaultReleaseTag,
		}
		incorrectName := "incorrect"
		mockUpgrader.EXPECT().GetReleaseWithAsset(meshctl.Ctx, opts.ReleaseTag, upgrade_assets.MeshctlBinaryName()).Return(&github.RepositoryRelease{
			TagName: github.String(cliconstants.DefaultReleaseTag),
			Assets: []github.ReleaseAsset{
				{
					Name: github.String(incorrectName),
				},
			},
		}, nil)
		_, err := meshctl.Invoke("upgrade")
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(upgrade.CouldNotFindAssetForReleaseError(upgrade_assets.MeshctlBinaryName(), opts.ReleaseTag)))
	})

	It("will error if asset cannot be downloaded", func() {
		opts := options.Upgrade{
			ReleaseTag:   cliconstants.DefaultReleaseTag,
			DownloadPath: "downloadPath",
		}
		browserDownloadUrl := "download-url"
		mockUpgrader.EXPECT().GetReleaseWithAsset(meshctl.Ctx, opts.ReleaseTag, upgrade_assets.MeshctlBinaryName()).Return(&github.RepositoryRelease{
			Assets: []github.ReleaseAsset{
				{
					Name:               github.String(upgrade_assets.MeshctlBinaryName()),
					BrowserDownloadURL: github.String(browserDownloadUrl),
				},
			},
		}, nil)
		mockUpgrader.EXPECT().DownloadAsset(browserDownloadUrl, opts.DownloadPath).Return(testErr)
		_, err := meshctl.Invoke(fmt.Sprintf("upgrade --path=%s", opts.DownloadPath))
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(upgrade.DownloadingAssetError(testErr, upgrade_assets.MeshctlBinaryName())))
	})

	var (
		tagName = "vtest"

		finalOutput = fmt.Sprintf(`Downloading %s from release tag %s
Successfully downloaded and installed meshctl version %s to`, upgrade_assets.MeshctlBinaryName(), tagName, tagName)
	)

	It("will return nil if asset is downloaded correctly", func() {
		opts := options.Upgrade{
			ReleaseTag:   cliconstants.DefaultReleaseTag,
			DownloadPath: "",
		}
		browserDownloadUrl := "download-url"
		mockUpgrader.EXPECT().GetReleaseWithAsset(meshctl.Ctx, opts.ReleaseTag, upgrade_assets.MeshctlBinaryName()).Return(&github.RepositoryRelease{
			TagName: github.String(tagName),
			Assets: []github.ReleaseAsset{
				{
					Name:               github.String(upgrade_assets.MeshctlBinaryName()),
					BrowserDownloadURL: github.String(browserDownloadUrl),
				},
			},
		}, nil)
		mockUpgrader.EXPECT().DownloadAsset(browserDownloadUrl, opts.DownloadPath).Return(nil)
		output, err := meshctl.Invoke("upgrade")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(finalOutput))
	})

})
