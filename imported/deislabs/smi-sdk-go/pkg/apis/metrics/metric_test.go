package metrics

import (
	"testing"

	assertLib "github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	assert := assertLib.New(t)

	m := &Metric{}
	val := 10.123456

	m.Set(val)

	assert.Equal("10123m", m.Value.String())
}
