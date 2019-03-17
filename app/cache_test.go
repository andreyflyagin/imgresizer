package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_add(t *testing.T) {
	c := newCache(10, cacheLiveTime)
	c.add("key1", []byte("hello"), "hash1")
	data, hash := c.get("key1")
	assert.Equal(t, "hello", string(data))
	assert.Equal(t, "hash1", hash)

	//evicting test
	c.add("key2", []byte("1234567"), "hash1")
	data, hash = c.get("key1")
	assert.Nil(t, data)
	assert.Equal(t, "", hash)
}

func Test_get(t *testing.T) {
	c := newCache(20, time.Millisecond * 50)
	c.add("key1", []byte("hello"), "hash1")
	data, hash := c.get("key1")
	assert.Equal(t, "hello", string(data))
	assert.Equal(t, "hash1", hash)

	data, hash = c.get("non exist")
	assert.Nil(t, data)
	assert.Equal(t, "", hash)

	// test cache live time
	time.Sleep(time.Millisecond * 60)

	data, hash = c.get("key1")
	assert.Nil(t, data)
	assert.Equal(t, "", hash)
}
