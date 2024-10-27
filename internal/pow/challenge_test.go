package pow

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestHashCash(t *testing.T) {
	bits := 4
	resource := "test"

	h := NewHashcash(bits, resource)
	assert.False(t, h.Check()) // challenge unresolved

	assert.Error(t, h.Compute(0)) // zero tries given

	assert.NoError(t, h.Compute(math.MaxInt))
	assert.True(t, h.Check())
}
