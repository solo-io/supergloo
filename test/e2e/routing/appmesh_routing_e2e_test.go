package routing

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsappmesh "github.com/aws/aws-sdk-go/service/appmesh"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/utils/nameutils"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/test/helpers"
	testsetup "github.com/solo-io/solo-kit/test/setup"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/setup"
	"github.com/solo-io/supergloo/test/utils"
)

var _ = Describe("appmesh routing E2e", func() {
	var namespace string

	BeforeEach(func() {
		namespace = "appmesh-routing-test-" + helpers.RandString(8)
		err := testsetup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {

		if false {
			// wait for sig usr1
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGUSR1)
			defer signal.Reset(syscall.SIGUSR1)
			fmt.Println("We are here:")
			debug.PrintStack()
			fmt.Printf("Waiting for human intervention. to continue, run 'kill -SIGUSR1 %d'\n", os.Getpid())
			<-c
		}

		gexec.TerminateAndWait(2 * time.Second)

		sess, err := session.NewSession(aws.NewConfig().
			WithCredentials(credentials.NewSharedCredentials("", "")))
		Expect(err).NotTo(HaveOccurred())
		appmeshClient := awsappmesh.New(sess, &aws.Config{Region: aws.String("us-east-1")})
		list, err := appmeshClient.ListMeshes(&awsappmesh.ListMeshesInput{})
		Expect(err).NotTo(HaveOccurred())
		for _, mesh := range list.Meshes {

			vnlist, err := appmeshClient.ListVirtualNodes(&awsappmesh.ListVirtualNodesInput{
				MeshName: mesh.MeshName,
			})
			Expect(err).NotTo(HaveOccurred())
			for _, vn := range vnlist.VirtualNodes {
				_, err := appmeshClient.DeleteVirtualNode(&awsappmesh.DeleteVirtualNodeInput{
					MeshName:        mesh.MeshName,
					VirtualNodeName: vn.VirtualNodeName,
				})
				Expect(err).NotTo(HaveOccurred())
			}

			vrlist, err := appmeshClient.ListVirtualRouters(&awsappmesh.ListVirtualRoutersInput{
				MeshName: mesh.MeshName,
			})
			Expect(err).NotTo(HaveOccurred())
			for _, vr := range vrlist.VirtualRouters {

				rlist, err := appmeshClient.ListRoutes(&awsappmesh.ListRoutesInput{
					VirtualRouterName: vr.VirtualRouterName,
					MeshName:          mesh.MeshName,
				})
				Expect(err).NotTo(HaveOccurred())
				for _, r := range rlist.Routes {
					_, err = appmeshClient.DeleteRoute(&awsappmesh.DeleteRouteInput{
						MeshName:          mesh.MeshName,
						VirtualRouterName: vr.VirtualRouterName,
						RouteName:         r.RouteName,
					})
					Expect(err).NotTo(HaveOccurred())
				}

				_, err = appmeshClient.DeleteVirtualRouter(&awsappmesh.DeleteVirtualRouterInput{
					MeshName:          mesh.MeshName,
					VirtualRouterName: vr.VirtualRouterName,
				})
				Expect(err).NotTo(HaveOccurred())
			}

			_, err = appmeshClient.DeleteMesh(&awsappmesh.DeleteMeshInput{
				MeshName: mesh.MeshName,
			})
			Expect(err).NotTo(HaveOccurred())
		}

		testsetup.TeardownKube(namespace)
	})

	It("works", func() {
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
		cmd := exec.Command(PathToUds, "--namespace="+namespace)
		cmd.Env = os.Environ()
		_, err := gexec.Start(cmd, os.Stdout, os.Stdout)
		Expect(err).NotTo(HaveOccurred())

		meshes, _, secretClient, err := v1Clients()
		Expect(err).NotTo(HaveOccurred())

		ref := utils.SetupAppMesh(meshes, secretClient, namespace)

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		err = utils.DeployBookinfoAppMesh(cfg, namespace, appmesh.MeshName(ref), "us-east-1")
		Expect(err).NotTo(HaveOccurred())
		testrunnerHost := "testrunner." + namespace + ".svc.cluster.local"
		testrunnerVirtualNodeName := nameutils.SanitizeName(testrunnerHost)
		err = utils.DeployTestRunnerAppMesh(cfg, namespace, appmesh.MeshName(ref), testrunnerVirtualNodeName, "us-east-1")
		Expect(err).NotTo(HaveOccurred())

		// TODO ilackarms: remove this code when the AWS Service Limits increase
		// for now we must delete some deployments
		kube, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		err = kube.ExtensionsV1beta1().Deployments(namespace).Delete("details-v1", nil)
		Expect(err).NotTo(HaveOccurred())
		err = kube.ExtensionsV1beta1().Deployments(namespace).Delete("ratings-v1", nil)
		Expect(err).NotTo(HaveOccurred())
		err = kube.ExtensionsV1beta1().Deployments(namespace).Delete("productpage-v1", nil)
		Expect(err).NotTo(HaveOccurred())
		err = kube.CoreV1().Services(namespace).Delete("details", nil)
		Expect(err).NotTo(HaveOccurred())
		err = kube.CoreV1().Services(namespace).Delete("ratings", nil)
		Expect(err).NotTo(HaveOccurred())
		err = kube.CoreV1().Services(namespace).Delete("productpage", nil)
		Expect(err).NotTo(HaveOccurred())

		// reviews v1
		Eventually(func() (string, error) {
			return testsetup.Curl(testsetup.CurlOpts{
				Method:  "GET",
				Path:    "/reviews/1",
				Service: "reviews." + namespace + ".svc.cluster.local",
				Port:    9080,
			})
		}, time.Second*120).Should(ContainSubstring(`{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!"},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare."}]}`))

	})
})

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
