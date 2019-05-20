package metrics

import (
	"strings"
	"testing"
	"time"

	assertLib "github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewResource(t *testing.T) {
	assert := assertLib.New(t)

	testCases := []struct {
		obj  *v1.ObjectReference
		edge *v1.ObjectReference
		has  []string
		len  int
	}{
		{
			&v1.ObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "foo",
				Namespace:  "bar",
			},
			nil,
			[]string{"deployments", "bar"},
			8,
		},
		{
			&v1.ObjectReference{
				Kind:      "Random",
				Namespace: "other",
			},
			nil,
			[]string{"unsupported", "other"},
			7,
		},
		{
			&v1.ObjectReference{
				Kind: "Random",
			},
			nil,
			[]string{"unsupported"},
			5,
		},
		{
			&v1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Namespace",
				Name:       "foo",
			},
			nil,
			[]string{"namespaces", "foo"},
			6,
		},
		{
			&v1.ObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "foo",
				Namespace:  "bar",
			},
			&v1.ObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "foo",
				Namespace:  "bar",
			},
			[]string{"deployments", "bar"},
			9,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.obj.Kind, func(t *testing.T) {
			r := NewTrafficMetrics(tc.obj, tc.edge)

			assert.Equal("TrafficMetrics", r.TypeMeta.Kind)
			assert.Equal(APIVersion, r.TypeMeta.APIVersion)

			assert.WithinDuration(
				time.Now(),
				r.ObjectMeta.CreationTimestamp.Time,
				10*time.Millisecond)
			for _, item := range tc.has {
				assert.Contains(r.ObjectMeta.SelfLink, item)
			}
			assert.Len(strings.Split(r.ObjectMeta.SelfLink, "/"), tc.len)
			assert.Equal(tc.obj.Name, r.ObjectMeta.Name)
			assert.Equal(tc.obj.Namespace, r.ObjectMeta.Namespace)

			assert.Equal(tc.obj, r.Resource)

			if tc.edge != nil {
				assert.Equal(tc.edge, r.Edge.Resource)
			}

			assert.Len(r.Metrics, 5)
			for _, item := range AvailableMetrics {
				assert.Equal(item, r.Get(item.Name))
			}

			assert.Nil(r.Get("foobar"))
		})
	}
}
