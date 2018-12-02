package routing

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsappmesh "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/solo-kit/pkg/utils/nameutils"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/test/helpers"
	testsetup "github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/setup"
	"github.com/solo-io/supergloo/test/utils"
)

var _ = FDescribe("appmesh routing E2e", func() {
	var namespace string

	BeforeEach(func() {
		namespace = "appmesh-routing-test-" + helpers.RandString(8)
		namespace = "supergloo-system"
		return
		err := testsetup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		gexec.TerminateAndWait(2 * time.Second)

		sess, err := session.NewSession(aws.NewConfig().
			WithCredentials(credentials.NewSharedCredentials("", "")))
		Expect(err).NotTo(HaveOccurred())
		appmeshClient := awsappmesh.New(sess, &aws.Config{Region: aws.String("us-east-1")})
		list, err := appmeshClient.ListMeshes(&awsappmesh.ListMeshesInput{})
		Expect(err).NotTo(HaveOccurred())
		for _, mesh := range list.Meshes {
			_, err := appmeshClient.DeleteMesh(&awsappmesh.DeleteMeshInput{
				MeshName: mesh.MeshName,
			})
			Expect(err).NotTo(HaveOccurred())
		}

		return
		testsetup.TeardownKube(namespace)
	})

	It("works", func() {
		if false {
			go setup.Main(func(e error) {
				defer GinkgoRecover()
				if e == nil {
					return
				}
				if strings.Contains(e.Error(), "upstream") {
					return
				}
				Fail(e.Error())
			}, namespace)

			// start discovery
			cmd := exec.Command(PathToUds, "--namespace", namespace)
			cmd.Env = os.Environ()
			_, err := gexec.Start(cmd, os.Stdout, os.Stdout)
			Expect(err).NotTo(HaveOccurred())
		}

		meshes, routingRules, secretClient, err := v1Clients()
		Expect(err).NotTo(HaveOccurred())

		ref := setupAppMesh(meshes, secretClient, namespace)

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		err = utils.DeployBookinfoAppMesh(cfg, namespace, appmesh.MeshName(ref), "us-east-1")
		Expect(err).NotTo(HaveOccurred())
		testrunnerHost := "testrunner." + namespace + ".svc.cluster.local"
		testrunnerVirtualNodeName := nameutils.SanitizeName(testrunnerHost)
		err = utils.DeployTestRunnerAppMesh(cfg, namespace, appmesh.MeshName(ref), testrunnerVirtualNodeName, "us-east-1")
		Expect(err).NotTo(HaveOccurred())

		setupV1RoutingRule(routingRules, namespace, &ref)

		// reviews v1
		Eventually(func() (string, error) {
			return testsetup.Curl(testsetup.CurlOpts{
				Method:  "GET",
				Path:    "/reviews/1",
				Service: "reviews",
				Port:    9080,
			})
		}, time.Second*120).Should(ContainSubstring(`{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!"},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare."}]}`))

		setupV2RoutingRule(routingRules, namespace, &ref)

		// reviews v2
		Eventually(func() (string, error) {
			return testsetup.Curl(testsetup.CurlOpts{
				Method:  "GET",
				Path:    "/reviews/1",
				Service: "reviews",
				Port:    9080,
			})
		}, time.Second*120).Should(ContainSubstring(`{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!", "rating": {"stars": 5, "color": "black"}},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.", "rating": {"stars": 4, "color": "black"}}]}`))

	})
})

func setupAppMesh(meshClient v1.MeshClient, secretClient gloov1.SecretClient, namespace string) core.ResourceRef {
	secretMeta := core.Metadata{Name: "my-appmesh-credentials", Namespace: namespace}
	creds, err := credentials.NewSharedCredentials("", "").Get()
	Expect(err).NotTo(HaveOccurred())
	secretClient.Delete(namespace, secretMeta.Name, clients.DeleteOpts{})
	secret1, err := secretClient.Write(&gloov1.Secret{
		Metadata: secretMeta,
		Kind: &gloov1.Secret_Aws{
			Aws: &gloov1.AwsSecret{
				// these can be read in from ~/.aws/credentials by default (if user does not provide)
				// see https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html for more details
				AccessKey: creds.AccessKeyID,
				SecretKey: creds.SecretAccessKey,
			},
		},
	}, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())

	meshMeta := core.Metadata{Name: "my-appmesh", Namespace: namespace}
	meshClient.Delete(meshMeta.Namespace, meshMeta.Name, clients.DeleteOpts{})

	ref := secret1.Metadata.Ref()
	mesh1, err := meshClient.Write(&v1.Mesh{
		Metadata: meshMeta,
		MeshType: &v1.Mesh_AppMesh{
			AppMesh: &v1.AppMesh{
				AwsRegion:      "us-east-1",
				AwsCredentials: &ref,
			},
		},
	}, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(mesh1).NotTo(BeNil())
	return mesh1.Metadata.Ref()
}

func v1Clients() (v1.MeshClient, v1.RoutingRuleClient, gloov1.SecretClient, error) {
	kubeCache := kube.NewKubeCache()
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, nil, nil, err
	}
	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if err := meshClient.Register(); err != nil {
		return nil, nil, nil, err
	}

	routingRuleClient, err := v1.NewRoutingRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1.RoutingRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if err := routingRuleClient.Register(); err != nil {
		return nil, nil, nil, err
	}

	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	secretClient, err := gloov1.NewSecretClient(&factory.KubeSecretClientFactory{
		Clientset: kube,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if err := secretClient.Register(); err != nil {
		return nil, nil, nil, err
	}

	return meshClient, routingRuleClient, secretClient, nil
}
