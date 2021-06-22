// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bugreport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kr/pretty"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/archive"
	cluster2 "github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/cluster"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/common"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/config"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/content"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/filter"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/kubectlcmd"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/report/internal/processlog"
	version2 "github.com/solo-io/gloo-mesh/pkg/meshctl/commands/version"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	analyzer_util "istio.io/istio/galley/pkg/config/analysis/analyzers/util"
	"istio.io/istio/operator/pkg/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/kube"
	"istio.io/istio/pkg/proxy"
	"istio.io/pkg/log"
)

const (
	bugReportDefaultTimeout = 30 * time.Minute
	istioRevisionLabel      = "istio.io/rev"
)

var (
	bugReportDefaultIstioNamespace    = "istio-system"
	bugReportDefaultGlooMeshNamespace = "gloo-mesh"
	bugReportDefaultInclude           = []string{""}
	bugReportDefaultExclude           = []string{strings.Join(analyzer_util.SystemNamespaces, ", ")}
)

// Cmd returns a cobra command for report.
func Cmd(logOpts *log.Options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "report",
		Short:        "meshctl debug report selectively captures cluster information and logs into an archive to help diagnose problems.",
		SilenceUsage: true,
		Long: `meshctl debug report selectively captures cluster information and logs into an archive to help diagnose problems.
Proxy logs can be filtered using:
  --include|--exclude ns1,ns2.../dep1,dep2.../pod1,pod2.../cntr1,cntr.../lbl1=val1,lbl2=val2.../ann1=val1,ann2=val2...
where ns=namespace, dep=deployment, cntr=container, lbl=label, ann=annotation

For multiple clusters use commas to separate kube configs or kube contexts
--kubeconfig ~/.kube/config --context cluster-1,cluster-2
--kubeconfig ~/.kube/cluster1,~/.kube/cluster2 --context cluster-1,cluster-2

The --kubeconfig and --context flags are not needed if you already have your meshctl config file set up (see meshctl cluster configure).

The filter spec is interpreted as 'must be in (ns1 OR ns2) AND (dep1 OR dep2) AND (cntr1 OR cntr2)...'
The log will be included only if the container matches at least one include filter and does not match any exclude filters.
All parts of the filter are optional and can be omitted e.g. ns1//pod1 filters only for namespace ns1 and pod1.
All names except label and annotation keys support '*' glob matching pattern.

e.g.
--include ns1,ns2 (only namespaces ns1 and ns2)
--include n*//p*/l=v* (pods with name beginning with 'p' in namespaces beginning with 'n' and having label 'l' with value beginning with 'v'.)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBugReportCommand(cmd, logOpts)
		},
	}
	addFlags(rootCmd, gConfig)

	return rootCmd
}

var (
	// Logs, along with stats and importance metrics. Key is filepath (namespace/deployment/pod/cluster) which can be
	// parsed with ParsePath.
	logs       = make(map[string]string)
	stats      = make(map[string]*processlog.Stats)
	importance = make(map[string]int)
	// Aggregated errors for all fetch operations.
	gErrors util.Errors
	lock    = sync.RWMutex{}
)

func runBugReportCommand(_ *cobra.Command, logOpts *log.Options) error {
	if err := configLogs(logOpts); err != nil {
		return err
	}
	config, err := parseConfig()
	if err != nil {
		return err
	}

	kubeConfigs := strings.Split(config.KubeConfigPath, ",")
	contexts := strings.Split(config.Context, ",")

	if len(kubeConfigs) > 1 && len(contexts) > 1 && len(kubeConfigs) != len(contexts) {
		return errors.New("cannot match kubeconfigs with contexts")
	}
	// establish rootdir, will create one if needed
	tempDir = archive.GetRootDir(tempDir)
	tempDirPlaceholder := tempDir

	currentClusterContext, err := content.GetClusterContext()
	if err == nil {
		common.LogAndPrintf("\nCurrent cluster context: %s\n", currentClusterContext)
	}
	combos := buildKubeConfigList(kubeConfigs, contexts, len(config.KubeConfigPath+config.Context) > 0)
	for name, meshctlCluster := range combos {
		kubeConfigPath := meshctlCluster.KubeConfig
		kubeContext := meshctlCluster.KubeContext
		common.LogAndPrintf("\nTarget cluster config: %s\n", kubeConfigPath)
		common.LogAndPrintf("Running with the following context: \n\n%s\n\n", kubeContext)

		// override tempdir per clusters
		tempDir = fmt.Sprintf("%s/%s", tempDirPlaceholder, name)

		clientConfig, clientset, err := utils.BuildClientConfigAndClientset(kubeConfigPath, kubeContext)
		if err != nil {
			return fmt.Errorf("could not initialize k8s client: %s ", err)
		}
		client, err := kube.NewExtendedClient(clientConfig, "")
		if err != nil {
			return err
		}
		resources, err := cluster2.GetClusterResources(context.Background(), clientset)
		if err != nil {
			return err
		}

		dumpGlooMeshVersions(kubeConfigPath, kubeContext, config.GlooMeshNamespace)
		dumpMeshctlCheck(kubeConfigPath, kubeContext, config.GlooMeshNamespace)
		dumpRevisionsAndVersions(resources, kubeConfigPath, kubeContext, config.IstioNamespace)

		log.Infof("Cluster resource tree:\n\n%s\n\n", resources)
		paths, err := filter.GetMatchingPaths(config, resources)
		if err != nil {
			return err
		}

		common.LogAndPrintf("\n\nFetching proxy logs for the following containers:\n\n%s\n", strings.Join(paths, "\n"))

		gatherInfo(client, config, resources, paths)
		if len(gErrors) != 0 {
			log.Error(gErrors.ToError())
		}

		// TODO: sort by importance and discard any over the size limit.
		for path, text := range logs {
			namespace, _, pod, _, err := cluster2.ParsePath(path)
			if err != nil {
				log.Errorf(err.Error())
				continue
			}
			writeFile(filepath.Join(archive.ProxyOutputPath(tempDir, namespace, pod), common.ProxyContainerName+".log"), text)
		}
	}

	// reset tempdir so archiving can happen correctly
	tempDir = tempDirPlaceholder

	outDir, err := os.Getwd()
	if err != nil {
		log.Errorf("using ./ to write archive: %s", err.Error())
		outDir = "."
	}
	outPath := filepath.Join(outDir, "meshctl-bug-report.tgz")
	common.LogAndPrintf("Creating an archive at %s.\n", outPath)

	archiveDir := archive.DirToArchive(tempDir)
	if err := archive.Create(archiveDir, outPath); err != nil {
		return err
	}
	common.LogAndPrintf("Cleaning up temporary files in %s.\n", archiveDir)
	if err := os.RemoveAll(archiveDir); err != nil {
		return err
	}
	common.LogAndPrintf("Done.\n")
	return nil
}

// If no flags were passed in or if a non-default meshctl config file path was passed in,
// use the meshctl config file.
// Otherwise, parse the combos from the following combinations
// no kubeconfig, list of contexts (meaning use default kubeconfig)
// all kubeconfigs, no contexts (meaning use default context in each kubeconfig)
// and equal list of kubeconfigs and contexts (use each corresponding context per kubeconfig)
func buildKubeConfigList(kubeconfigs, kubecontexts []string, useFlags bool) map[string]utils.MeshctlCluster {
	if !useFlags || gConfig.ConfigFilePath != utils.DefaultConfigPath {
		if clusters, err := utils.ParseMeshctlConfig(gConfig.ConfigFilePath); err == nil {
			return clusters.Clusters
		}
	}

	combos := make(map[string]utils.MeshctlCluster)
	var clusterN int
	switch len(kubeconfigs) {
	case 1:
		{
			// all kubeconfigs, no contexts (meaning use default context in each kubeconfig)
			if len(kubecontexts) == 1 {
				// 1 to 1 kubeconfig to context
				combos[fmt.Sprintf("cluster-%d", clusterN)] = utils.MeshctlCluster{
					KubeConfig:  kubeconfigs[0],
					KubeContext: kubecontexts[0],
				}
				clusterN++
			} else {
				// return the single kubeconfig
				for _, c := range kubecontexts {
					combos[fmt.Sprintf("cluster-%s", clusterN)] = utils.MeshctlCluster{
						KubeConfig:  kubeconfigs[0],
						KubeContext: c,
					}
					clusterN++
				}
			}
		}
	default:
		// all kubeconfigs, no contexts (meaning use default context in each kubeconfig)
		if len(kubecontexts) == 1 {
			for _, k := range kubeconfigs {
				combos[fmt.Sprintf("cluster-%s", clusterN)] = utils.MeshctlCluster{
					KubeConfig:  k,
					KubeContext: "",
				}
				clusterN++
			}
		} else {
			// and equal list of kubeconfigs and contexts (use each corresponding context per kubeconfig)
			for i, c := range kubecontexts {
				combos[fmt.Sprintf("cluster-%s", clusterN)] = utils.MeshctlCluster{
					KubeConfig:  kubeconfigs[i],
					KubeContext: c,
				}
				clusterN++
			}
		}
	}
	return combos
}

func dumpMeshctlCheck(kubeconfig, context, glooMeshNamespace string) {
	var b bytes.Buffer
	utils.RunShell(fmt.Sprintf("meshctl check --kubeconfig \"%s\" --kubecontext \"%s\" --namespace \"%s\"",
		kubeconfig, context, glooMeshNamespace), io.Writer(&b))
	text := b.String()
	common.LogAndPrintf(text)
	appendToFile(filepath.Join(archive.GlooMeshPath(tempDir), "meshctl-check"), text)
}

func dumpGlooMeshVersions(kubeconfig, context, glooMeshNamespace string) {
	text := getGlooMeshVersion(kubeconfig, context, glooMeshNamespace)

	common.LogAndPrintf(text)
	appendToFile(filepath.Join(archive.GlooMeshPath(tempDir), "gloo-versions"), text)
}

func getGlooMeshVersion(kubeconfig, configcontext, glooMeshNamespace string) string {
	text := ""
	glooMeshVersions := version2.MakeServerVersions(context.Background(), &version2.Options{
		Kubeconfig:  kubeconfig,
		Kubecontext: configcontext,
		Namespace:   glooMeshNamespace,
	})
	text += "The following Gloo versions were found in the cluster\nKubeconfig: " + kubeconfig + "\nContext: " + configcontext + "\n"
	for _, ver := range glooMeshVersions {
		text += fmt.Sprintf("Namespace: %s\n", ver.Namespace)
		for _, c := range ver.Components {
			text += fmt.Sprintf("\tName: %s\n", c.ComponentName)
			for _, i := range c.Images {
				text += fmt.Sprintf("\t\tName: %s Image: %s/%s:%s\n", i.Name, i.Domain, i.Path, i.Version)
			}
		}
	}
	return text
}

func dumpRevisionsAndVersions(resources *cluster2.Resources, kubeconfig, configContext, istioNamespace string) {
	text := ""
	revisions := getIstioRevisions(resources)
	istioVersions, proxyVersions := getIstioVersions(kubeconfig, configContext, istioNamespace, revisions)
	text += "The following Istio control plane revisions/versions were found in the cluster:\n"
	for rev, ver := range istioVersions {
		text += fmt.Sprintf("Revision %s:\n%s\n\n", rev, ver)
	}
	text += "The following proxy revisions/versions were found in the cluster:\n"
	for rev, ver := range proxyVersions {
		text += fmt.Sprintf("Revision %s: Versions {%s}\n", rev, strings.Join(ver, ", "))
	}
	common.LogAndPrintf(text)
	writeFile(filepath.Join(archive.OutputRootDir(tempDir), "versions"), text)
}

// getIstioRevisions returns a slice with all Istio revisions detected in the cluster.
func getIstioRevisions(resources *cluster2.Resources) []string {
	revMap := make(map[string]struct{})
	for _, podLabels := range resources.Labels {
		for label, value := range podLabels {
			if label == istioRevisionLabel {
				revMap[value] = struct{}{}
			}
		}
	}
	var out []string
	for k := range revMap {
		out = append(out, k)
	}
	return out
}

// getIstioVersions returns a mapping of revision to aggregated version string for Istio components and revision to
// slice of versions for proxies. Any errors are embedded in the revision strings.
func getIstioVersions(kubeconfig, configContext, istioNamespace string, revisions []string) (map[string]string, map[string][]string) {
	istioVersions := make(map[string]string)
	proxyVersionsMap := make(map[string]map[string]struct{})
	proxyVersions := make(map[string][]string)
	for _, revision := range revisions {
		istioVersions[revision] = getIstioVersion(kubeconfig, configContext, istioNamespace, revision)
		proxyInfo, err := proxy.GetProxyInfo(kubeconfig, configContext, revision, istioNamespace)
		if err != nil {
			log.Error(err)
			continue
		}
		for _, pi := range *proxyInfo {
			if proxyVersionsMap[revision] == nil {
				proxyVersionsMap[revision] = make(map[string]struct{})
			}
			proxyVersionsMap[revision][pi.IstioVersion] = struct{}{}
		}
	}
	for revision, vmap := range proxyVersionsMap {
		for version := range vmap {
			proxyVersions[revision] = append(proxyVersions[revision], version)
		}
	}
	return istioVersions, proxyVersions
}

func getIstioVersion(kubeconfig, configContext, istioNamespace, revision string) string {
	kubeClient, err := kube.NewExtendedClient(kube.BuildClientCmd(kubeconfig, configContext), revision)
	if err != nil {
		return err.Error()
	}

	versions, err := kubeClient.GetIstioVersions(context.TODO(), istioNamespace)
	if err != nil {
		return err.Error()
	}
	return pretty.Sprint(versions)
}

// gatherInfo fetches all logs, resources, debug etc. using goroutines.
// proxy logs and info are saved in logs/stats/importance global maps.
// Errors are reported through gErrors.
func gatherInfo(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources, paths []string) {
	// no timeout on mandatoryWg.
	var mandatoryWg sync.WaitGroup
	cmdTimer := time.NewTimer(time.Duration(config.CommandTimeout))

	clusterDir := archive.ClusterInfoPath(tempDir)

	params := &content.Params{
		Client: client,
		DryRun: config.DryRun,
	}
	common.LogAndPrintf("\nFetching Istio control plane information from cluster.\n\n")
	getFromCluster(content.GetK8sResources, params, clusterDir, &mandatoryWg)
	getFromCluster(content.GetCRs, params, clusterDir, &mandatoryWg)
	getFromCluster(content.GetEvents, params, clusterDir, &mandatoryWg)
	getFromCluster(content.GetClusterInfo, params, clusterDir, &mandatoryWg)
	getFromCluster(content.GetSecrets, params.SetVerbose(config.FullSecrets), clusterDir, &mandatoryWg)
	getFromCluster(content.GetDescribePods, params.SetIstioNamespace(config.IstioNamespace), clusterDir, &mandatoryWg)
	// Gloo mesh describe pods
	getFromCluster(content.GetDescribePods, params.SetIstioNamespace(config.GlooMeshNamespace), archive.GlooMeshPath(tempDir), &mandatoryWg)

	// optionalWg is subject to timer.
	var optionalWg sync.WaitGroup
	for _, p := range paths {
		namespace, _, pod, container, err := cluster2.ParsePath(p)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		cp := params.SetNamespace(namespace).SetPod(pod).SetContainer(container)
		proxyDir := archive.ProxyOutputPath(tempDir, namespace, pod)
		switch {
		case common.IsProxyContainer(params.ClusterVersion, container):
			getFromCluster(content.GetCoredumps, cp, filepath.Join(proxyDir, "cores"), &mandatoryWg)
			getFromCluster(content.GetNetstat, cp, proxyDir, &mandatoryWg)
			getFromCluster(content.GetProxyInfo, cp, archive.ProxyOutputPath(tempDir, namespace, pod), &optionalWg)
			getProxyLogs(client, config, resources, p, namespace, pod, container, &optionalWg)

		case resources.IsDiscoveryContainer(params.ClusterVersion, namespace, pod, container):
			getFromCluster(content.GetIstiodInfo, cp, archive.IstiodPath(tempDir, namespace, pod), &mandatoryWg)
			getIstiodLogs(client, config, resources, namespace, pod, &mandatoryWg)
		case resources.IsGlooMeshDashboardContainer(container):
			getGlooMeshDashboardLogs(client, config, resources, namespace, pod, &mandatoryWg)
		case resources.IsGlooMeshAgentContainer(container):
			getGlooMeshAgentLogs(client, config, resources, namespace, pod, &mandatoryWg)
		case resources.IsGlooMeshEnterpriseNetworkingContainer(container):
			getGlooMeshEnterpriseNetworkingLogs(client, config, resources, namespace, pod, &mandatoryWg)
		case resources.IsGlooMeshDiscoveryContainer(container):
			getGlooMeshDiscoveryLogs(client, config, resources, namespace, pod, &mandatoryWg)
		case resources.IsGlooMeshNetworkingContainer(container):
			getGlooMeshNetworkingLogs(client, config, resources, namespace, pod, &mandatoryWg)
		case common.IsOperatorContainer(params.ClusterVersion, container):
			getOperatorLogs(client, config, resources, namespace, pod, &optionalWg)
		}
	}

	// Not all items are subject to timeout. Proceed only if the non-cancellable items have completed.
	mandatoryWg.Wait()

	// If log fetches have completed, cancel the timeout.
	go func() {
		optionalWg.Wait()
		cmdTimer.Reset(0)
	}()

	// Wait for log fetches, up to the timeout.
	<-cmdTimer.C

	// Analyze runs many queries internally, so run these queries sequentially and after everything else has finished.
	runAnalyze(config, resources, params)
}

// getFromCluster runs a cluster info fetching function f against the cluster and writes the results to fileName.
// Runs if a goroutine, with errors reported through gErrors.
func getFromCluster(f func(params *content.Params) (map[string]string, error), params *content.Params, dir string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on %s", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
	go func() {
		defer wg.Done()
		out, err := f(params)
		appendGlobalErr(err)
		if err == nil {
			writeFiles(dir, out)
		}
		log.Infof("Done with %s", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
	}()
}

// getProxyLogs fetches proxy logs for the given namespace/pod/container and stores the output in global structs.
// Runs if a goroutine, with errors reported through gErrors.
// TODO(stewartbutler): output the logs to a more robust/complete structure.
func getProxyLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	path, namespace, pod, container string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, cstat, imp, err := getLog(client, resources, config, namespace, pod, container)
		appendGlobalErr(err)
		lock.Lock()
		if err == nil {
			logs[path], stats[path], importance[path] = clog, cstat, imp
		}
		lock.Unlock()
		log.Infof("Done with logs %s", pod)
	}()
}

// getIstiodLogs fetches Istiod logs for the given namespace/pod and writes the output.
// Runs if a goroutine, with errors reported through gErrors.
func getIstiodLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	namespace, pod string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, _, _, err := getLog(client, resources, config, namespace, pod, common.DiscoveryContainerName)
		appendGlobalErr(err)
		writeFile(filepath.Join(archive.IstiodPath(tempDir, namespace, pod), "discovery.log"), clog)
		log.Infof("Done with logs %s", pod)
	}()
}

// getGlooMeshAgentLogs fetches Gloo mesh logs for the given namespace/pod and writes the output.
// Runs if a goroutine, with errors reported through gErrors.
func getGlooMeshAgentLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	namespace, pod string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, err := getGlooLog(client, resources, config, namespace, pod, cluster2.GlooMeshEnterpriseAgentContainerName)
		appendGlobalErr(err)
		writeFile(filepath.Join(archive.GlooMeshPath(tempDir), "enterprise-agent.log"), clog)
		log.Infof("Done with logs %s", pod)
	}()
}

// getGlooMeshEnterpriseNetworkingLogs fetches Gloo mesh logs for the given namespace/pod and writes the output.
// Runs if a goroutine, with errors reported through gErrors.
func getGlooMeshEnterpriseNetworkingLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	namespace, pod string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, err := getGlooLog(client, resources, config, namespace, pod, cluster2.GlooMeshEnterpriseNetworkingContainerName)
		appendGlobalErr(err)
		writeFile(filepath.Join(archive.GlooMeshPath(tempDir), "enterprise-networking.log"), clog)
		log.Infof("Done with logs %s", pod)
	}()
}

// getGlooMeshEnterpriseNetworkingLogs fetches Gloo mesh logs for the given namespace/pod and writes the output.
// Runs if a goroutine, with errors reported through gErrors.
func getGlooMeshDashboardLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	namespace, pod string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, err := getGlooLog(client, resources, config, namespace, pod, cluster2.GlooMeshDashboardContainerName)
		appendGlobalErr(err)
		writeFile(filepath.Join(archive.GlooMeshPath(tempDir), "dashboard.log"), clog)
		log.Infof("Done with logs %s", pod)
	}()
}

// getGlooMeshDiscoveryLogs fetches Gloo mesh logs for the given namespace/pod and writes the output.
// Runs if a goroutine, with errors reported through gErrors.
func getGlooMeshDiscoveryLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	namespace, pod string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, err := getGlooLog(client, resources, config, namespace, pod, cluster2.GlooMeshDiscoveryContainerName)
		appendGlobalErr(err)
		writeFile(filepath.Join(archive.GlooMeshPath(tempDir), "dashboard.log"), clog)
		log.Infof("Done with logs %s", pod)
	}()
}

// getGlooMeshNetworkingLogs fetches Gloo mesh logs for the given namespace/pod and writes the output.
// Runs if a goroutine, with errors reported through gErrors.
func getGlooMeshNetworkingLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	namespace, pod string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, err := getGlooLog(client, resources, config, namespace, pod, cluster2.GlooMeshNetworkingContainerName)
		appendGlobalErr(err)
		writeFile(filepath.Join(archive.GlooMeshPath(tempDir), "dashboard.log"), clog)
		log.Infof("Done with logs %s", pod)
	}()
}

// getOperatorLogs fetches istio-operator logs for the given namespace/pod and writes the output.
func getOperatorLogs(client kube.ExtendedClient, config *config.BugReportConfig, resources *cluster2.Resources,
	namespace, pod string, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Infof("Waiting on logs %s", pod)
	go func() {
		defer wg.Done()
		clog, _, _, err := getLog(client, resources, config, namespace, pod, common.OperatorContainerName)
		appendGlobalErr(err)
		writeFile(filepath.Join(archive.OperatorPath(tempDir, namespace, pod), "operator.log"), clog)
		log.Infof("Done with logs %s", pod)
	}()
}

// getLog fetches the logs for the given namespace/pod/container and returns the log text and stats for it.
func getLog(client kube.ExtendedClient, resources *cluster2.Resources, config *config.BugReportConfig,
	namespace, pod, container string) (string, *processlog.Stats, int, error) {
	log.Infof("Getting logs for %s/%s/%s...", namespace, pod, container)
	clog, err := kubectlcmd.Logs(client, namespace, pod, container, false, config.DryRun)
	if err != nil {
		return "", nil, 0, err
	}
	if resources.ContainerRestarts(namespace, pod, container) > 0 {
		pclog, err := kubectlcmd.Logs(client, namespace, pod, container, true, config.DryRun)
		if err != nil {
			return "", nil, 0, err
		}
		clog = "========= Previous log present (appended at the end) =========\n\n" + clog +
			"\n\n========= Previous log =========\n\n" + pclog
	}
	var cstat *processlog.Stats
	clog, cstat = processlog.Process(config, clog)
	return clog, cstat, cstat.Importance(), nil
}

// getGlooLog fetches the logs for the given namespace/pod/container and returns the log text.
func getGlooLog(client kube.ExtendedClient, resources *cluster2.Resources, config *config.BugReportConfig,
	namespace, pod, container string) (string, error) {
	log.Infof("Getting logs for %s/%s/%s...", namespace, pod, container)
	clog, err := kubectlcmd.Logs(client, namespace, pod, container, false, config.DryRun)
	if err != nil {
		return "", err
	}
	if resources.ContainerRestarts(namespace, pod, container) > 0 {
		pclog, err := kubectlcmd.Logs(client, namespace, pod, container, true, config.DryRun)
		if err != nil {
			return "", err
		}
		clog = "========= Previous log present (appended at the end) =========\n\n" + clog +
			"\n\n========= Previous log =========\n\n" + pclog
	}
	return clog, nil
}

func runAnalyze(config *config.BugReportConfig, resources *cluster2.Resources, params *content.Params) {
	for ns := range resources.Root {
		if analyzer_util.IsSystemNamespace(resource.Namespace(ns)) {
			continue
		}
		common.LogAndPrintf("Running istio analyze on namespace %s.\n", ns)
		out, err := content.GetAnalyze(params.SetIstioNamespace(config.IstioNamespace))
		if err != nil {
			log.Error(err.Error())
			continue
		}
		writeFiles(archive.AnalyzePath(tempDir, ns), out)
	}
	common.LogAndPrintf("\n")
}

func writeFiles(dir string, files map[string]string) {
	for fname, text := range files {
		writeFile(filepath.Join(dir, fname), text)
	}
}

func writeFile(path, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	mkdirOrExit(path)
	if err := ioutil.WriteFile(path, []byte(text), 0644); err != nil {
		log.Errorf(err.Error())
	}
}

func appendToFile(path, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	mkdirOrExit(path)
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf(err.Error())
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		log.Errorf(err.Error())
	}
}

func mkdirOrExit(fpath string) {
	if err := os.MkdirAll(path.Dir(fpath), 0755); err != nil {
		fmt.Printf("Could not create output directories: %s", err)
		os.Exit(-1)
	}
}

func appendGlobalErr(err error) {
	if err == nil {
		return
	}
	lock.Lock()
	gErrors = util.AppendErr(gErrors, err)
	lock.Unlock()
}

func configLogs(opt *log.Options) error {
	logDir := filepath.Join(archive.OutputRootDir(tempDir), "meshctl-bug-report.log")
	mkdirOrExit(logDir)
	f, err := os.Create(logDir)
	if err != nil {
		return err
	}
	f.Close()
	op := []string{logDir}
	opt2 := *opt
	opt2.OutputPaths = op
	opt2.ErrorOutputPaths = op
	opt2.SetOutputLevel("default", log.InfoLevel)

	return log.Configure(&opt2)
}
