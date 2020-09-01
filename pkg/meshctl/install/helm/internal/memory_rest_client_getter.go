package internal

// Copied from https://github.com/solo-io/go-utils/blob/2af7be12c6cfd50ed5890f59fbf03b4b404c5707/installutils/helminstall/internal/memory_backed_rest_client_getter.go#L1

import (
	"path/filepath"
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	defaultCacheDir = filepath.Join(clientcmd.RecommendedConfigDir, "http-cache")

	// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
	overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)

	removeProtocol = regexp.MustCompile("http[s]?:/{2}")
)

func NewInMemoryRESTClientGetter(rawConfig clientcmd.ClientConfig) genericclioptions.RESTClientGetter {
	return &inMemoryRESTClientGetter{rawConfig: rawConfig}
}

type inMemoryRESTClientGetter struct {
	rawConfig clientcmd.ClientConfig
}

func (i *inMemoryRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	config, err := i.rawConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (i *inMemoryRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	cfg, err := i.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return ToDiscoveryClient(cfg)
}

func (i *inMemoryRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	cfg, err := i.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return ToRESTMapper(cfg)
}

func (i *inMemoryRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return i.rawConfig
}

//
// files below this comment were shamelessly and brazenly stolen from the Helm codebase
//

func ToDiscoveryClient(restCfg *rest.Config) (discovery.CachedDiscoveryInterface, error) {
	discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube", "cache", "discovery"), restCfg.Host)
	return diskcached.NewCachedDiscoveryClientForConfig(restCfg, discoveryCacheDir, defaultCacheDir, 10*time.Minute)
}

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := removeProtocol.ReplaceAllString(host, "")
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
}

func ToRESTMapper(restCfg *rest.Config) (meta.RESTMapper, error) {
	discoveryClient, err := ToDiscoveryClient(restCfg)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}
