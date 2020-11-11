package printing

import (
	"fmt"
	"strings"

	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
)

func FormattedField(key, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%s: %s\n", key, value)
}

func FormattedObjectRef(ref *v1.ObjectRef) string {
	if ref == nil {
		return ""
	}
	return FormattedObjectRefs([]*v1.ObjectRef{ref})
}

func FormattedObjectRefs(refs []*v1.ObjectRef) string {
	if len(refs) < 1 {
		return ""
	}
	var s strings.Builder
	for i, ref := range refs {
		s.WriteString(FormattedField("Name", ref.Name))
		s.WriteString(FormattedField("Namespace", ref.Namespace))
		if i < len(refs)-1 {
			s.WriteString("\n")
		}
	}
	return s.String()
}

func FormattedClusterObjectRef(ref *v1.ClusterObjectRef) string {
	if ref == nil {
		return ""
	}
	return FormattedClusterObjectRefs([]*v1.ClusterObjectRef{ref})
}

func FormattedClusterObjectRefs(refs []*v1.ClusterObjectRef) string {
	if len(refs) < 1 {
		return ""
	}
	var s strings.Builder
	for i, ref := range refs {
		s.WriteString(FormattedField("Name", ref.Name))
		s.WriteString(FormattedField("Namespace", ref.Namespace))
		s.WriteString(FormattedField("Cluster", ref.ClusterName))
		if i < len(refs)-1 {
			s.WriteString("\n")
		}
	}
	return s.String()
}
