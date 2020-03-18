package healthcheck_types

import "sort"

// visit each category in the health check suite in weight order (higher weighted categories first)
func (h HealthCheckSuite) ForEachCategoryInWeightOrder(f func(category Category)) {
	var sortedCategories SortableCategories
	for category, _ := range h {
		sortedCategories = append(sortedCategories, category)
	}

	sort.Sort(sortedCategories)

	for _, categoryInWeightOrder := range sortedCategories {
		f(categoryInWeightOrder)
	}
}
