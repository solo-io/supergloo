package helm

import (
	"os"

	"github.com/solo-io/service-mesh-hub/codegen/io"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"

	"github.com/solo-io/skv2/codegen/model"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// this file provides the source of truth for the Service Mesh Hub Helm chart.
// it is imported into the root level generate.go to generate the Portal manifest

var (
	registry = func() string {
		registry := os.Getenv("IMAGE_REGISTRY")
		if registry == "" {
			registry = defaultRegistry
		}
		return registry
	}()

	// built-in skv2 templates we don't need
	filterTemplates = func(outPath string) bool {
		return outPath == "templates/namespace.yaml" ||
			outPath == "templates/chart.yaml"
	}
)

var Chart = &model.Chart{
	Operators: []model.Operator{
		discoveryOperator(),
		networkingOperator(),
	},
	FilterTemplate: filterTemplates,
	Data: model.Data{
		ApiVersion:  "v1",
		Name:        "service-mesh-hub",
		Description: "Helm chart for Service Mesh Hub.",
		Version:     version.Version,
	},
	Values: defaultValues(),
}

var CertAgentChart = &model.Chart{
	Operators: []model.Operator{
		certAgentOperator(),
	},
	FilterTemplate: filterTemplates,
	Data: model.Data{
		ApiVersion:  "v1",
		Name:        "cert-agent",
		Description: "Helm chart for the Service Mesh Hub Certificate Agent.",
		Version:     version.Version,
	},
	Values: nil,
}

var (
	defaultRegistry = "soloio"
)

func discoveryOperator() model.Operator {

	var rbacPolicies []rbacv1.PolicyRule

	rbacPolicies = append(rbacPolicies, io.ClusterWatcherInputTypes.RbacPoliciesWatch()...)
	rbacPolicies = append(rbacPolicies, io.DiscoveryOutputTypes.Snapshot.RbacPoliciesWrite()...)

	return model.Operator{
		Name: "discovery",
		Deployment: model.Deployment{
			Image: smhImage(),
			Resources: &v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("125m"),
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
		},
		Service: model.Service{
			Type: v1.ServiceTypeClusterIP,
			Ports: []model.ServicePort{
				{
					Name:        "metrics",
					DefaultPort: 9091,
				},
			},
		},
		Rbac: rbacPolicies,
		Args: []string{
			"discovery",
			"--metrics-port={{ $.Values.discovery.ports.metrics }}",
			"--verbose",
		},
		Env: []v1.EnvVar{
			{
				Name: defaults.PodNamespaceEnv,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
	}
}

func networkingOperator() model.Operator {

	var rbacPolicies []rbacv1.PolicyRule

	rbacPolicies = append(rbacPolicies, io.ClusterWatcherInputTypes.RbacPoliciesWatch()...)
	rbacPolicies = append(rbacPolicies, io.NetworkingInputTypes.RbacPoliciesWatch()...)
	rbacPolicies = append(rbacPolicies, io.NetworkingInputTypes.RbacPoliciesUpdateStatus()...)
	rbacPolicies = append(rbacPolicies, io.LocalNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	rbacPolicies = append(rbacPolicies, io.IstioNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	rbacPolicies = append(rbacPolicies, io.SmiNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	rbacPolicies = append(rbacPolicies, io.CertificateIssuerInputTypes.RbacPoliciesWatch()...)
	rbacPolicies = append(rbacPolicies, io.CertificateIssuerInputTypes.RbacPoliciesUpdateStatus()...)

	return model.Operator{
		Name: "networking",
		Deployment: model.Deployment{
			Image: smhImage(),
			Resources: &v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("125m"),
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
		},
		Service: model.Service{
			Type: v1.ServiceTypeClusterIP,
			Ports: []model.ServicePort{
				{
					Name:        "metrics",
					DefaultPort: 9091,
				},
			},
		},
		Rbac: rbacPolicies,
		Args: []string{
			"networking",
			"--metrics-port={{ $.Values.networking.ports.metrics }}",
			"--verbose",
		},
		Env: []v1.EnvVar{
			{
				Name: defaults.PodNamespaceEnv,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
	}
}

func certAgentOperator() model.Operator {

	var rbacPolicies []rbacv1.PolicyRule

	rbacPolicies = append(rbacPolicies, io.CertificateAgentInputTypes.RbacPoliciesWatch()...)
	rbacPolicies = append(rbacPolicies, io.CertificateAgentInputTypes.RbacPoliciesUpdateStatus()...)
	rbacPolicies = append(rbacPolicies, io.CertificateAgentOutputTypes.Snapshot.RbacPoliciesWrite()...)
	// ability to bounce pods
	rbacPolicies = append(rbacPolicies, rbacv1.PolicyRule{
		Verbs:     []string{"*"},
		APIGroups: []string{""},
		Resources: []string{"pods"},
	})

	return model.Operator{
		Name: "cert-agent",
		Deployment: model.Deployment{
			Image: certAgentImage(),
			Resources: &v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("50m"),
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		Service: model.Service{
			Type: v1.ServiceTypeClusterIP,
			Ports: []model.ServicePort{
				{
					Name:        "metrics",
					DefaultPort: 9091,
				},
			},
		},
		Rbac: rbacPolicies,
		Args: []string{
			"--metrics-port={{ $.Values.certAgent.ports.metrics }}",
			"--verbose",
		},
		Env: []v1.EnvVar{
			{
				Name: defaults.PodNamespaceEnv,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
	}
}

// both smh operators share same image
func smhImage() model.Image {
	return model.Image{
		Registry:   registry,
		Repository: "service-mesh-hub",
		Tag:        version.Version,
		PullPolicy: v1.PullIfNotPresent,
	}
}

func certAgentImage() model.Image {
	return model.Image{
		Registry:   registry,
		Repository: "cert-agent",
		Tag:        version.Version,
		PullPolicy: v1.PullIfNotPresent,
	}
}
