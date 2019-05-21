package metrics

import (
	"testing"

	assertLib "github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewList(t *testing.T) {
	assert := assertLib.New(t)

	testCases := []struct {
		obj  *v1.ObjectReference
		edges bool
		link string
	}{
		{
			&v1.ObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "foo",
				Namespace:  "bar",
			},
			false,
			"deployments",
		},
		{
			&v1.ObjectReference{
				Kind: "Random",
			},
			false,
			"unsupported",
		},
		{
			&v1.ObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "foo",
				Namespace:  "bar",
			},
			true,
			"namespaces/bar/deployments/foo/edges",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.obj.Kind, func(t *testing.T) {
			lst := NewTrafficMetricsList(tc.obj, tc.edges)

			assert.Equal("TrafficMetricsList", lst.TypeMeta.Kind)
			assert.Equal(APIVersion, lst.TypeMeta.APIVersion)
			assert.Contains(lst.SelfLink, tc.link)
			assert.Equal(tc.obj, lst.Resource)

			assert.Len(lst.Items, 0)
		})
	}
}

func TestListGet(t *testing.T) {
	assert := assertLib.New(t)

	lst := NewTrafficMetricsList(&v1.ObjectReference{}, false)

	assert.Len(lst.Items, 0)

	obj := &v1.ObjectReference{
		Kind:      "deployment",
		Namespace: "default",
		Name:      "foo",
	}

	assert.Equal(obj, lst.Get(obj, nil).Resource)
	assert.Len(lst.Items, 1)

	assert.Equal(obj, lst.Get(obj, nil).Resource)
	assert.Len(lst.Items, 1)

	assert.Equal(obj, lst.Get(obj, obj).Edge.Resource)
	assert.Len(lst.Items, 2)

	assert.Equal(obj, lst.Get(obj, obj).Edge.Resource)
	assert.Len(lst.Items, 2)
}
