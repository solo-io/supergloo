package printing

import (
	"fmt"
	"strings"

	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
)

type DescriptionBuilder struct {
	s strings.Builder
}

func NewDescriptionBuilder() *DescriptionBuilder {
	return &DescriptionBuilder{}
}

func (d *DescriptionBuilder) AddHeader(indent int, header string) {
	d.s.WriteString(indentedLine(indent, fmt.Sprintf("%s:", header)))
}

func (d *DescriptionBuilder) AddField(indent int, key, value string) {
	if value == "" {
		return
	}
	d.s.WriteString(indentedLine(indent, fmt.Sprintf("%s: %s", key, value)))
}

func (d *DescriptionBuilder) AddObjectRefs(indent int, header string, refs []*v1.ObjectRef){
	if len(refs) < 1 {
		return
	}
	d.s.WriteString("\n")
	d.s.WriteString(FormattedHeader(indent, header))
	indent += 2
	for i, ref := range refs {
		d.s.WriteString(FormattedField(indent, "Name", ref.Name))
		d.s.WriteString(FormattedField(indent, "Namespace", ref.Namespace))
		if i < len(refs)-1 {
			d.s.WriteString("\n")
		}
	}
}

func (d *DescriptionBuilder) String() string {
	return d.s.String()
}

func FormattedHeader(indent int, header string) string {
	return indentedLine(indent, fmt.Sprintf("%s:", header))
}

func FormattedField(indent int, key, value string) string {
	if value == "" {
		return ""
	}
	return indentedLine(indent, fmt.Sprintf("%s: %s", key, value))
}

func FormattedObjectRefs(indent int, header string, refs []*v1.ObjectRef) string {
	if len(refs) < 1 {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(FormattedHeader(indent, header))
	indent += 2
	for i, ref := range refs {
		s.WriteString(FormattedField(indent, "Name", ref.Name))
		s.WriteString(FormattedField(indent, "Namespace", ref.Namespace))
		if i < len(refs)-1 {
			s.WriteString("\n")
		}
	}
	return s.String()
}

func indentedLine(level int, s string) string {
	indent := strings.Repeat(" ", level)
	return fmt.Sprintf("%s%s\n", indent, s)
}
