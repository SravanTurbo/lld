package lrucache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestCapacityZero(t *testing.T){
	assert.Panics(t, func() {NewLRUCache(0)})
}

func TestGetRecencyUpdate(t *testing.T){
	lc := NewLRUCache(2)
	lc.Put("a", "1")
	lc.Put("b", "1")
	lc.Get("a")

	assert.Equal(t, lc.head.next.key, "a")
	
}

func TestEviction(t *testing.T){
	lc := NewLRUCache(2)
	lc.Put("a", "1")
	lc.Put("b", "1")
	lc.Put("c", "1")

	assert.Equal(t, len(lc.cache), 2)
}