package int_set

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntSet(t *testing.T) {
	s := new(IntSet)
	assert.True(t, s.Empty())
	err := s.Set("123,456-789")
	require.NoError(t, err)
	assert.False(t, s.Empty())
	assert.False(t, s.Contains(122))
	assert.True(t, s.Contains(123))
	assert.False(t, s.Contains(124))

	assert.False(t, s.Contains(455))
	for i := 456; i <= 789; i++ {
		assert.True(t, s.Contains(i))
	}
	assert.False(t, s.Contains(790))
}
