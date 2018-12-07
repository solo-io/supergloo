package cli_test

import (
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

var _ = Describe("cli e2e", func() {

	type CliTest struct {
		name string
		args []string
	}

	It("Root command", func() {
		recorder := &Recorder{}
		cmd := exec.Command(SuperglooExec)
		cmd.Stdout = recorder
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred())
		Expect(recorder.data).To(Equal(utils.RootCommandResponse))
	})

	//It("Init", func() {
	//	cmd := exec.Command(SuperglooExec, "init")
	//	err := cmd.Run()
	//	Expect(err).NotTo(HaveOccurred())
	//})

	Describe("Install/Uninstall", func() {

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

		routingRuleTests := []CliTest{
			{
				name: "fault-injection",
				args: []string{},
			},
		}

		var recorder *Recorder

		BeforeEach(func() {
			recorder = &Recorder{}
		})

		cleanUpName := func(name string) string {
			name = strings.Replace(name, "-", " ", -1)
			name = strings.Title(name)
			return name
		}

		for _, test := range routingRuleTests {
			It(cleanUpName(test.name), func() {
				cmd := exec.Command(SuperglooExec, test.args...)
				cmd.Stdout = recorder
				err := cmd.Run()
				Expect(err).NotTo(HaveOccurred())
			})
		}
	})
})
