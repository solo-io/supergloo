package metrics

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rotisserie/eris"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/k8s-utils/debugutils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// filters for snapshots
	networking           = "networking"
	discovery            = "discovery"
	enterpriseNetworking = "enterprise-networking"
	enterpriseAgent      = "enterprise-agent"
)

type DebugMetricsOpts struct {
	file    string
	zip     string
	dir     string
	verbose bool

	kubeconfig  string
	kubecontext string

	// hidden optional values
	metricsBindPort uint32
	namespace       string
}

func AddDebugMetricsFlags(flags *pflag.FlagSet, opts *DebugMetricsOpts) {
	utils.AddManagementKubeconfigFlags(&opts.kubeconfig, &opts.kubecontext, flags)
	flags.StringVarP(&opts.file, "file", "f", "", "file to write output to")
	flags.StringVar(&opts.dir, "dir", "", "dir to write file outputs to")
	flags.StringVar(&opts.zip, "zip", "", "zip file output")
	flags.Uint32Var(&opts.metricsBindPort, "port", defaults.MetricsPort, "metrics port")
	flags.StringVarP(&opts.namespace, "namespace", "n", defaults.GetPodNamespace(), "gloo-mesh namespace")
}

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &DebugMetricsOpts{}

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "metrics for the discovery and networking pods.",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.verbose = globalFlags.Verbose
			return debugMetrics(ctx, opts, []string{discovery, networking, enterpriseNetworking, enterpriseAgent})
		},
	}
	cmd.AddCommand(
		Networking(ctx, opts),
		Discovery(ctx, opts),
		EnterpriseNetworking(ctx, opts),
		EnterpriseAgent(ctx, opts),
	)
	AddDebugMetricsFlags(cmd.PersistentFlags(), opts)

	cmd.PersistentFlags().Lookup("namespace").Hidden = true
	cmd.PersistentFlags().Lookup("port").Hidden = true
	return cmd
}

func EnterpriseNetworking(ctx context.Context, opts *DebugMetricsOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enterprise-networking",
		Short: "Input and output snapshots for the enterprise networking pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugMetrics(ctx, opts, []string{enterpriseNetworking})
		},
	}
	return cmd
}

func EnterpriseAgent(ctx context.Context, opts *DebugMetricsOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enterprise-agent",
		Short: "Input and output snapshots for the enterprise agent pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugMetrics(ctx, opts, []string{enterpriseAgent})
		},
	}
	return cmd
}

func Networking(ctx context.Context, opts *DebugMetricsOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "networking",
		Short: "Input and output snapshots for the networking pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugMetrics(ctx, opts, []string{networking})
		},
	}
	return cmd
}

func Discovery(ctx context.Context, opts *DebugMetricsOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Input and output snapshots for the discovery pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugMetrics(ctx, opts, []string{discovery})
		},
	}
	return cmd
}

func debugMetrics(ctx context.Context, opts *DebugMetricsOpts, pods []string) error {
	f, err := os.Create(opts.file)
	defer f.Close()

	fs := afero.NewOsFs()
	zipDir, err := afero.TempDir(fs, "", "")
	if err != nil {
		return err
	}
	defer fs.RemoveAll(zipDir)
	storageClient := debugutils.NewFileStorageClient(fs)
	for _, podName := range pods {
		fmt.Printf("Metrics for %s:\n", podName)
		metrics, metricsErr := getMetrics(ctx, opts, "", podName)
		if metricsErr != nil {
			fmt.Println(metricsErr.Error())
			continue
		}
		fileName := fmt.Sprintf("%s-metrics.json", podName)
		if opts.file != "" {
			_, err = f.WriteString(metrics)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			fmt.Printf("Written to %s\n", opts.file)
		} else if opts.zip != "" || opts.dir != "" {
			if len(metrics) == 0 {
				continue
			}
			dir := zipDir
			if opts.dir != "" {
				dir = opts.dir
			}
			err = storageClient.Save(dir, &debugutils.StorageObject{
				Resource: strings.NewReader(metrics),
				Name:     fileName,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Written to %s\n", fileName)
		} else {
			fmt.Print(metrics)
			fmt.Print("\n")
		}
	}
	if opts.zip != "" {
		err = utils.Zip(fs, zipDir, opts.zip)
	}
	return nil
}

func getMetrics(ctx context.Context, opts *DebugMetricsOpts, localPort, podName string) (string, error) {
	kubeClient, err := utils.BuildClientset(opts.kubeconfig, opts.kubecontext)
	if err != nil {
		return "", err
	}
	_, err = kubeClient.AppsV1().Deployments(opts.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", eris.Errorf("No %s.%s deployment found\n", opts.namespace, podName)
	}

	portFwdContext, cancelPtFwd := context.WithCancel(ctx)
	mgmtDeployNamespace := opts.namespace
	mgmtDeployName := podName
	remotePort := strconv.Itoa(int(opts.metricsBindPort))
	// start port forward to mgmt server stats port
	localPort, err = utils.PortForwardFromDeployment(
		portFwdContext,
		opts.kubeconfig,
		opts.kubecontext,
		mgmtDeployName,
		mgmtDeployNamespace,
		fmt.Sprintf("%v", localPort),
		fmt.Sprintf("%v", remotePort),
	)
	if err != nil {
		return "", eris.Errorf("try verifying that `kubectl port-forward -n %v deployment/%v %v:%v` can be run successfully.", mgmtDeployNamespace, mgmtDeployName, localPort, remotePort)
	}
	// request metrics page
	metricsUrl := fmt.Sprintf("http://localhost:%v/metrics", localPort)
	resp, err := http.DefaultClient.Get(metricsUrl)
	if err != nil {
		return "", eris.Errorf("try verifying that the %s pod is listening on port %v", podName, remotePort)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	snapshot := string(b)

	cancelPtFwd()

	return snapshot, nil
}
