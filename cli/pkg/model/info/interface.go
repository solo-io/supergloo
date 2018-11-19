package info

type ResourceInfo interface {
	// returns the already formatted Header row
	Headers() string
	// returns a slice where each element is a formatted resource row
	Resources() []string
}

type Header struct {
	// Name that will be displayed when printing to terminal
	Name string

	// If true, this field will be displayed only when using the "-o wide" option
	WideOnly bool
}

// For each row, we will want to look up the values by the header name
type Data []map[string]Field

type Field struct {
	Header Header
	Value  string
}

func (h Header) String() string {
	return h.Name
}
