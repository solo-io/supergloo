package kuberesource

import (
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/helm/pkg/tiller"
)

// set the install order based on
// Helm's list
var installOrder = func() map[string]int {
	installOrderMapping := make(map[string]int)
	for installValue, kind := range tiller.InstallOrder {
		installOrderMapping[kind] = installValue + 1 // add 1, the 0 value is unknown type
	}
	return installOrderMapping
}()

func getInstallOrder(res *unstructured.Unstructured) (int, bool) {
	kind := res.GroupVersionKind().Kind

	order, listed := installOrder[kind]
	return order, listed
}

func sortUnstructured(resources UnstructuredResources) {
	sort.SliceStable(resources, func(i, j int) bool {
		res1 := resources[i]
		res2 := resources[j]

		installOrder1, listed1 := getInstallOrder(res1)
		installOrder2, listed2 := getInstallOrder(res2)
		kind1, kind2 := res1.GroupVersionKind().Kind, res2.GroupVersionKind().Kind

		// unlisted objects come last
		switch {
		case !listed1 && !listed2:
			if kind1 != kind2 {
				return kind1 < kind2
			}
			return nameLess(res1, res2)
		case !listed1:
			return false
		case !listed2:
			return true
		}

		if installOrder1 != installOrder2 {
			return installOrder1 < installOrder2
		}

		if kind1 != kind2 {
			return kind1 < kind2
		}

		// sort by namespace.name
		val := nameLess(res1, res2)
		return val
	})
}

func nameLess(res1, res2 *unstructured.Unstructured) bool {
	return res1.GetNamespace()+res1.GetName() < res2.GetNamespace()+res2.GetName()
}

func installOrderLess(kind1, kind2 string) bool {
	order1, listed1 := installOrder[kind1]
	order2, listed2 := installOrder[kind2]
	switch {
	case !listed1 && !listed2:
		return true
	case !listed1:
		return false
	case !listed2:
		return true
	}

	return order1 < order2
}
