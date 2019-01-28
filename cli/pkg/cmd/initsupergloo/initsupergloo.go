package initsupergloo

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/solo-io/supergloo/cli/pkg/cliconstants"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/cli/pkg/setup"
	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/spf13/cobra"

	kubecore "k8s.io/api/core/v1"
	kubeextv1b1 "k8s.io/api/extensions/v1beta1"
	kuberbacv1b1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultSuperGlooImage = "soloio/supergloo:" + cliconstants.SuperGlooImageTag
	defaultDiscoveryImage = "soloio/discovery:" + cliconstants.DiscoveryImageTag
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: `Initialize supergloo`,
		Long:  `Initialize supergloo`,
		Run: func(c *cobra.Command, args []string) {
			initSuperGloo(opts)
		},
	}
	iop := &opts.Init
	pflags := cmd.PersistentFlags()
	pflags.BoolVarP(&opts.Init.CustomInstall, cliconstants.CustomInstallFlag, "c", false,
		"Enable SuperGloo initialization using custom images. Use with \"--supergloo-image\" +" +
		"and/or \"--discovery-image\" flags. Defaults to false.")
	pflags.StringVar(&iop.SuperGlooImage, cliconstants.SuperGlooImageFlag, "",
		"Name of the Supergloo container image to use. Defaults to " + defaultSuperGlooImage)
	pflags.StringVar(&iop.DiscoveryImage, cliconstants.DiscoveryImageFlag, "",
		"Name of the Supergloo container image to use. Defaults to " + defaultDiscoveryImage)
	return cmd
}

// Check if supergloo is running on the cluster and deploy it if it isn't
func initSuperGloo(opts *options.Options) {
	checkOptions(opts)
	initHelm(opts)
	if opts.Init.CustomInstall {
		initSuperGlooCustom(opts)
	} else {
		kubectl := "kubectl"
		args := []string{"apply", "-f", "https://raw.githubusercontent.com/solo-io/supergloo/master/hack/install/supergloo.yaml"}

		cmd := exec.Command(kubectl, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			return
		}
	}

	// wait for supergloo pods to be ready
	if !setup.LoopUntilAllPodsReadyOrTimeout(constants.SuperglooNamespace, opts.Cache.KubeClient) {
		fmt.Println("Supergloo pods did not initialize.")
		os.Exit(1)
	}
	fmt.Printf("Supergloo is ready on kubernetes cluster.\n")
}

func checkOptions(opts *options.Options) {
	if opts.Init.CustomInstall && (opts.Init.SuperGlooImage == "" || opts.Init.DiscoveryImage == "") {
		fmt.Println("\"--supergloo-image\" or \"--discovery-image\" flags must be set when \"--custom-install\" " +
			"flag is set.")
		os.Exit(1)
	}
	if !opts.Init.CustomInstall && (opts.Init.SuperGlooImage != "" || opts.Init.DiscoveryImage != "") {
		fmt.Println("\"--custom-install\" flag must be set when \"--supergloo-image\" or \"--discovery-image\" " +
			"flags are set.")
		os.Exit(1)
	}
}

func initSuperGlooCustom(opts *options.Options) {
	createKubeNamespace(opts, constants.GlooNamespace)
	createKubeNamespace(opts, constants.SuperglooNamespace)
	createKubeClusterRoleBindings(opts, constants.SuperglooClusterRoleBindings)
	createKubeDeployment(opts, "supergloo")
	createKubeDeployment(opts, "discovery")
}

func initHelm(opts *options.Options) {
	initKubeCache(opts)
	if !setup.PodAppears("kube-system", opts.Cache.KubeClient, "tiller") {
		fmt.Printf("Ensuring helm is initialized on kubernetes cluster.\n")
		cmd := exec.Command("kubectl", "apply", "-f", common.HelmSetupFileName)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error initializing Helm: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Running helm init.\n")
		cmd = exec.Command("helm", "init", "--service-account", "tiller", "--upgrade")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running \"helm init\" command: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Waiting for Tiller pod to be ready.\n")
		if !setup.LoopUntilPodAppears("kube-system", opts.Cache.KubeClient, "tiller") {
			fmt.Println("Tiller pod didn't get created")
			os.Exit(1)
		}
		if !setup.LoopUntilAllPodsReadyOrTimeout("kube-system", opts.Cache.KubeClient, "tiller") {
			fmt.Println("Tiller pod was not ready.")
			os.Exit(1)
		}
		fmt.Printf("Helm is initialzed.\n")
	}
	return
}

func initKubeCache(opts *options.Options) {
	// Should never happen, since InitCache gets called first, but just in case
	if opts.Cache.KubeClient == nil {
		if err := setup.InitCache(opts); err != nil {
			fmt.Printf("Error initializing Kubernetes cache: %v", err)
			os.Exit(1)
		}
	}
	return
}

func createKubeNamespace(opts *options.Options, name string) {
	ns, err := opts.Cache.KubeClient.CoreV1().Namespaces().Get(name, kubemeta.GetOptions{})
	switch {
	case ns.Name == name:
		fmt.Printf("namespace %s already exists.\n", name)
		return
	case errors.IsNotFound(err):
		opts.Cache.KubeClient.CoreV1().Namespaces().Create(&kubecore.Namespace{
			ObjectMeta: kubemeta.ObjectMeta{
				Name: name,
			},
		})
		fmt.Printf("created namespace: %s.\n", name)
	default:
		fmt.Printf("error creating namespace %s. exiting supergloo initialization: %v", name, err)
		os.Exit(1)
	}
}

func createKubeClusterRoleBindings(opts *options.Options, name string) {
	crb, err := opts.Cache.KubeClient.RbacV1beta1().ClusterRoleBindings().Get(name, kubemeta.GetOptions{})
	switch {
	case crb.Name == name:
		fmt.Printf("cluster role binding %s already exists.\n", name)
		return
	case errors.IsNotFound(err):
		opts.Cache.KubeClient.RbacV1beta1().ClusterRoleBindings().Create(&kuberbacv1b1.ClusterRoleBinding{
			TypeMeta: kubemeta.TypeMeta{},
			ObjectMeta: kubemeta.ObjectMeta{
				Name: name,
			},
			Subjects: []kuberbacv1b1.Subject{
				kuberbacv1b1.Subject{
					Kind:      "ServiceAccount",
					Name:      "default",
					Namespace: constants.SuperglooNamespace,
				},
			},
			RoleRef: kuberbacv1b1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			},
		})
		fmt.Printf("created cluster role binding: %s.\n", name)
	default:
		fmt.Printf("error creating cluster role binding %s. exiting supergloo initialization: %v", crb, err)
		os.Exit(1)
	}
}

func createKubeDeployment(opts *options.Options, name string) {
	deploy, err := opts.Cache.KubeClient.ExtensionsV1beta1().Deployments(constants.SuperglooNamespace).Get(name,kubemeta.GetOptions{})
	switch {
	case deploy.Name == name:
		fmt.Printf("deployment %s already exists in namespace %s.\n", name, constants.SuperglooNamespace)
		return
	case errors.IsNotFound(err):
		labels := map[string]string{"gloo": name}
		var args []string
		var image string
		switch {
		case name == "supergloo":
			if opts.Init.SuperGlooImage != "" {
				image = opts.Init.SuperGlooImage
			} else {
				image = defaultSuperGlooImage
			}
		case name == "discovery":
			if opts.Init.DiscoveryImage != "" {
				image = opts.Init.DiscoveryImage
			} else {
				image = defaultDiscoveryImage
			}
			discoveryNamespaceFlag := "--namespace=" + constants.SuperglooNamespace
			args = append(args, discoveryNamespaceFlag)
		default:
			fmt.Printf("error creating deployment %s. only \"discovery\" and \"supergloo\" are supported deployment names.", name)
			os.Exit(1)
		}
		opts.Cache.KubeClient.ExtensionsV1beta1().Deployments(constants.SuperglooNamespace).Create(&kubeextv1b1.Deployment{
			TypeMeta: kubemeta.TypeMeta{},
			ObjectMeta: kubemeta.ObjectMeta{
				Name: name,
				Namespace: constants.SuperglooNamespace,
				Labels: labels,
			},
			Spec: kubeextv1b1.DeploymentSpec{
				Template: kubecore.PodTemplateSpec{
					ObjectMeta: kubemeta.ObjectMeta{
						Labels: labels,
					},
					Spec: kubecore.PodSpec{
						Containers: []kubecore.Container{
							kubecore.Container{
								Name: name,
								Image: image,
								ImagePullPolicy: "Always",
								Args: args,
							},
						},
					},
				},
			},
		})
		fmt.Printf("created deployment %s in namespace %s.\n", name, constants.SuperglooNamespace)
	default:
		fmt.Printf("error creating deployment %s in namespace %s. exiting supergloo initialization: %v", name, constants.SuperglooNamespace, err)
		os.Exit(1)
	}
}