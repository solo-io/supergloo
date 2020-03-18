package healthcheck_types

import (
	"sort"
)

// an array of categories that can have sort.Sort() called on it to produce an array of HIGHER weighted categories to LOWER weighted categories
type SortableCategories []Category

var _ sort.Interface = SortableCategories{}

func (s SortableCategories) Len() int {
	return len(s)
}

// want higher weight to lower weight, so swap the comparison operator
func (s SortableCategories) Less(i, j int) bool {
	return s[i].Weight > s[j].Weight
}

// Swap swaps the elements with indexes i and j.
func (s SortableCategories) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
