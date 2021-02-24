package gloomesh

const (
	GlooMeshRepoURI                    = "https://storage.googleapis.com/gloo-mesh"
	GlooMeshChartUriTemplate           = GlooMeshRepoURI + "/gloo-mesh/gloo-mesh-%s.tgz"
	AgentCrdsChartUriTemplate          = GlooMeshRepoURI + "/agent-crds/agent-crds-%s.tgz"
	CertAgentChartUriTemplate          = GlooMeshRepoURI + "/cert-agent/cert-agent-%s.tgz"
	GlooMeshEnterpriseRepoURI          = "https://storage.googleapis.com/gloo-mesh-enterprise"
	GlooMeshEnterpriseChartUriTemplate = GlooMeshEnterpriseRepoURI + "/gloo-mesh-enterprise/gloo-mesh-enterprise-%s.tgz"
	EnterpriseAgentChartUriTemplate    = GlooMeshEnterpriseRepoURI + "/enterprise-agent/enterprise-agent-%s.tgz"
	GlooMeshReleaseName                = "gloo-mesh"
	GlooMeshEnterpriseReleaseName      = "gloo-mesh-enterprise"
	AgentCrdsReleaseName               = "agent-crds"
	CertAgentReleaseName               = "cert-agent"
	EnterpriseAgentReleaseName         = "enterprise-agent"
)
