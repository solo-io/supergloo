package inputs

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func BookInfoSecurityRules(namespace string, targetMesh *core.ResourceRef) v1.SecurityRuleList {
	return v1.SecurityRuleList{
		{
			Metadata: core.Metadata{
				Name:      "allow-productpage-to-details",
				Namespace: namespace,
			},
			TargetMesh: targetMesh,
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_ServiceSelector_{
					ServiceSelector: &v1.PodSelector_ServiceSelector{
						Services: []core.ResourceRef{
							{Name: "productpage", Namespace: namespace},
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_ServiceSelector_{
					ServiceSelector: &v1.PodSelector_ServiceSelector{
						Services: []core.ResourceRef{
							{Name: "details", Namespace: namespace},
						},
					},
				},
			},
			AllowedMethods: []string{"GET"},
			AllowedPaths:   []string{"/details/*"},
		},
		{
			Metadata: core.Metadata{
				Name:      "allow-productpage-to-ratings",
				Namespace: namespace,
			},
			TargetMesh: targetMesh,
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_ServiceSelector_{
					ServiceSelector: &v1.PodSelector_ServiceSelector{
						Services: []core.ResourceRef{
							{Name: "productpage", Namespace: namespace},
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_ServiceSelector_{
					ServiceSelector: &v1.PodSelector_ServiceSelector{
						Services: []core.ResourceRef{
							{Name: "ratings", Namespace: namespace},
						},
					},
				},
			},
			AllowedMethods: []string{"GET"},
			AllowedPaths:   []string{"/ratings/*"},
		},
		{
			Metadata: core.Metadata{
				Name:      "allow-ratings-to-reviews",
				Namespace: namespace,
			},
			TargetMesh: targetMesh,
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_ServiceSelector_{
					ServiceSelector: &v1.PodSelector_ServiceSelector{
						Services: []core.ResourceRef{
							{Name: "ratings", Namespace: namespace},
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_ServiceSelector_{
					ServiceSelector: &v1.PodSelector_ServiceSelector{
						Services: []core.ResourceRef{
							{Name: "reviews", Namespace: namespace},
						},
					},
				},
			},
			AllowedMethods: []string{"GET"},
			AllowedPaths:   []string{"/reviews/*"},
		},
	}
}
