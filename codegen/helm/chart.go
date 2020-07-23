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

// this file provides the source of truth for the DevPortal Helm chart.
// it is imported into the root level generate.go to generate the Portal manifest

var Chart = &model.Chart{
	Operators: []model.Operator{
		discoveryOperator(),
		networkingOperator(),
	},
	// Exclude the standard namespace template
	FilterTemplate: func(outPath string) bool {
		return outPath == "templates/namespace.yaml"
	},
	Data: model.Data{
		ApiVersion:  "v1",
		Name:        "service-mesh-hub",
		Description: "Helm chart for Service Mesh Hub.",
		Version:     version.Version,
	},
	Values: defaultValues(),
}

var (
	defaultRegistry = "soloio"
)

func discoveryOperator() model.Operator {

	var rbacPolicies []rbacv1.PolicyRule

	rbacPolicies = append(rbacPolicies, io.ClusterWatcherInputTypes.RbacPoliciesWatch()...)
	rbacPolicies = append(rbacPolicies, io.DiscoveryOutputTypes.RbacPoliciesWrite()...)

	return model.Operator{
		Name: "discovery",
		Deployment: model.Deployment{
			Image: makeImage(),
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
			"--metrics-port={{ $.Values.networking.ports.metrics }}",
			"--verbose",
		},
		Env: []v1.EnvVar{
			{
				Name: defaults.GetPodNamespace(),
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
	rbacPolicies = append(rbacPolicies, io.NetworkingOutputIstioTypes.RbacPoliciesWrite()...)

	return model.Operator{
		Name: "networking",
		Deployment: model.Deployment{
			Image: makeImage(),
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
				Name: defaults.GetPodNamespace(),
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
	}
}

// cache and operator share same image
func makeImage() model.Image {
	registry := os.Getenv("IMAGE_REGISTRY")
	if registry == "" {
		registry = defaultRegistry
	}
	return model.Image{
		Registry:   registry,
		Repository: "service-mesh-hub",
		Tag:        version.Version,
		PullPolicy: v1.PullIfNotPresent,
	}
}
