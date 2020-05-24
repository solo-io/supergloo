package internal

import (
	"path/filepath"
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
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
