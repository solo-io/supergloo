package istio

import (
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/destinationrule"
	drplugin "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/destinationrule/plugin"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/virtualservice"
	vsplugin "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/virtualservice/plugin"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
)

// the dependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our dependencyFactory anywhere else
type dependencyFactory interface {
	makeVirtualServiceTranslator(clusters skv1alpha1sets.KubernetesClusterSet) virtualservice.Translator

	makeDestinationRuleTranslator(clusters skv1alpha1sets.KubernetesClusterSet) destinationrule.Translator
}

type dependencyFactoryImpl struct{}

func newDependencyFactory() dependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) makeVirtualServiceTranslator(clusters skv1alpha1sets.KubernetesClusterSet) virtualservice.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	pluginFactory := vsplugin.NewFactory()

	return virtualservice.NewTranslator(clusterDomains, pluginFactory)
}

func (d dependencyFactoryImpl) makeDestinationRuleTranslator(clusters skv1alpha1sets.KubernetesClusterSet) destinationrule.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	pluginFactory := drplugin.NewFactory()

	return destinationrule.NewTranslator(clusterDomains, pluginFactory)
}
