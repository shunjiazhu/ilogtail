package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericPool(t *testing.T) {
	bufferPool := NewGenericPool(func() []byte {
		return make([]byte, 0, 128)
	})
	b := bufferPool.Get()
	assert.Empty(t, b)
	*b = append(*b, 'a')
	assert.Equal(t, []byte{'a'}, *b)
	bufferPool.Put(b)
	b = bufferPool.Get()
	assert.Empty(t, b)
	bufferPool.Put(b)

	indexPool := NewGenericPool(func() []string {
		return make([]string, 0, 10)
	})
	i := indexPool.Get()
	assert.Empty(t, i)
	*i = append(*i, "a")
	assert.Equal(t, []string{"a"}, *i)
	indexPool.Put(i)
	i = indexPool.Get()
	assert.Empty(t, i)
	indexPool.Put(i)
}
