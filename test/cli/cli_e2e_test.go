package cli_test

import (
	"fmt"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/constants"
	"github.com/solo-io/supergloo/test/util"

	"os/exec"
	"strings"

	"github.com/solo-io/supergloo/test/utils"
)

/*
End to end tests for supergloo-cli
*/

type Recorder struct {
	data string
}

func (mw *Recorder) Write(p []byte) (n int, err error) {
	mw.data += string(p)
	return len(p), nil
}

type ArgsBuilder struct {
	RuleType string
	MeshName string
	Upstream string
	Args     string
}

func (b *ArgsBuilder) Write(p []byte) (n int, err error) {
	b.Args += string(p)
	return len(p), nil
}

func (b *ArgsBuilder) constructArgs() []string {
	argsString := utils.RoutingRulesArgs + b.Args
	tmpl, err := template.New(b.RuleType + "-template").Parse(argsString)
	Expect(err).NotTo(HaveOccurred())
	b.Args = ""
	err = tmpl.Execute(b, b)
	b.Args = strings.Replace(b.Args, "\n", "", -1)
	b.Args = strings.Replace(b.Args, "\t", " ", -1)
	untrimmedArgs := strings.Split(b.Args, " ")
	trimmedArgs := make([]string, 0)
	for _, v := range untrimmedArgs {
		if v != "" && v != " " {
			trimmedArgs = append(trimmedArgs, v)
		}
	}
	return trimmedArgs
}

var _ = Describe("cli e2e", func() {
	defer GinkgoRecover()

	type CliTest struct {
		name string
		args []string
	}

	kubecahe := kube.NewKubeCache()

	PIt("Root command", func() {
		recorder := &Recorder{}
		cmd := exec.Command(SuperglooExec)
		cmd.Stdout = recorder
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred())
		Expect(recorder.data).To(Equal(utils.RootCommandResponse))
	})

	XIt("Init", func() {
		cmd := exec.Command(SuperglooExec, "init")
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred())
	})

	PDescribe("Install/Uninstall", func() {

		var meshName string

		setMeshName := func(message, substr string) {
			name := strings.Replace(message, substr, "", -1)
			meshName = strings.TrimSuffix(strings.TrimSpace(name), ".")
		}

		It("Install", func() {
			successMessage := "Installing istio in namespace istio-system with name"
			args := []string{"install", "-m", "istio", "-n", "istio-system", "-s"}
			recorder := &Recorder{}
			cmd := exec.Command(SuperglooExec, args...)
			cmd.Stdout = recorder
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			Expect(recorder.data).To(ContainSubstring(successMessage))
			setMeshName(recorder.data, successMessage)
		})

		It("Uninstall", func() {
			successMessage := "Successfully uninstalled mesh"
			args := []string{"uninstall", "-s", "-n", meshName}
			recorder := &Recorder{}
			cmd := exec.Command(SuperglooExec, args...)
			cmd.Stdout = recorder
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			Expect(recorder.data).To(ContainSubstring(successMessage))
		})
	})

	Describe("Routing rules", func() {

		var (
			recorder          *Recorder
			meshes, upstreams resources.ResourceList
		)

		routingRuleTests := []ArgsBuilder{
			{
				RuleType: "fault-injection",
				Args:     utils.FaultInjectionArgs,
			},
			{
				RuleType: "cors",
				Args:     utils.CorsArgs,
			},
			{
				RuleType: "header-manipulation",
				Args:     utils.HeaderManipulationArgs,
			},
			{
				RuleType: "mirror",
				Args:     utils.MirrorArgs,
			},
			{
				RuleType: "retries",
				Args:     utils.RetriesArgs,
			},
			{
				RuleType: "timeout",
				Args:     utils.TimeoutArgs,
			},
			{
				RuleType: "traffic-shifting",
				Args:     utils.TrafficShiftingArgs,
			},
		}

		cleanUpName := func(name string) string {
			name = strings.Replace(name, "-", " ", -1)
			name = strings.Title(name)
			return name
		}

		BeforeEach(func() {
			recorder = &Recorder{data: ""}

			upstreamClient := util.GetUpstreamClient(kubecahe)
			meshClient := util.GetMeshClient(kubecahe)
			upstreams = util.GetResourceList(upstreamClient.BaseClient(), constants.GlooNamespace)
			meshes = util.GetResourceList(meshClient.BaseClient(), constants.SuperglooNamespace)
		})

		getResourceName := func(resources resources.ResourceList) core.ResourceRef {
			Expect(len(resources)).To(BeNumerically(">", 0))
			names := resources.Names()
			namespaces := resources.Namespaces()
			return core.ResourceRef{
				Namespace: namespaces[0],
				Name:      names[0],
			}
		}

		for _, test := range routingRuleTests {
			// TODO (EItanya): Figure out why this line is necessary
			localTest := test
			It(cleanUpName(localTest.RuleType), func() {
				us := getResourceName(upstreams)
				mesh := getResourceName(meshes)
				localTest.Upstream = fmt.Sprintf("%s:%s", us.Namespace, us.Name)
				localTest.MeshName = mesh.Name
				cmd := exec.Command(SuperglooExec, localTest.constructArgs()...)
				cmd.Stdout = recorder
				cmd.Stderr = recorder
				err := cmd.Run()
				if err != nil {
					fmt.Println(recorder.data)
				}
				Expect(err).NotTo(HaveOccurred())
			})

		}

	})
})
