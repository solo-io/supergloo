package info

import "github.com/solo-io/supergloo/cli/pkg/cmd/options"

type ResourceInfo interface {
	// Returns a slice containing header names
	Headers(opts options.Get) []string
	// Returns a slice of slices, each one containing the fields corresponding to the headers
	Resources(opts options.Get) [][]string
}

type Header struct {
	// Name that will be displayed when printing to terminal
	Name string

	// If true, this field will be displayed only when using the "-o wide" option
	WideOnly bool
}

// For each row, we will want to look up the values by the header name
type Data []map[string]string

func (h Header) String() string {
	return h.Name
}
