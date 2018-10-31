package cache_test

import (
    "fmt"
    "time"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/helloworld/redis-proxy/cache"
)

var (
    capacity = 10
    globalExpiry = 100 // 100 ms
)

func TestNewCache(t *testing.T) {
    c := cache.New(capacity, globalExpiry)
    assert.Equal(t, globalExpiry, int(c.GlobalExpiry), "incorrect GlobalExpiry")
    assert.Equal(t, capacity, c.Cache.MaxEntries, "incorrect Capacity")
}

func TestGetNonExistentKey(t *testing.T) {
    c := cache.New(capacity, globalExpiry)
    value, ok := c.Get("nonexistent")
    assert.Equal(t, false, ok, "returned found for non-existent key")
    assert.Equal(t, "", value, "returned non-empty value for non-existent key")
}

func TestAddToCache(t *testing.T) {
    c := cache.New(capacity, globalExpiry)
    c.Set("key", "value")

    value, ok := c.Get("key")
    assert.Equal(t, true, ok, "returned not-found for existing key")
    assert.Equal(t, "value", value, "value incorrectly set")
}

func TestRemoveFromCache(t *testing.T) {
    c := cache.New(capacity, globalExpiry)
    c.Set("key", "value")
    c.Remove("key")

    value, ok := c.Get("key")
    assert.Equal(t, false, ok, "returned found for removed key")
    assert.Equal(t, "", value, "returned non-empty value for removed key")
}

func TestExpiry(t *testing.T) {
    c := cache.New(capacity, globalExpiry)
    c.Set("key", "value")

    time.Sleep(101 * time.Millisecond)
    value, ok := c.Get("key")

    assert.Equal(t, false, ok, "returned found for expired key")
    assert.Equal(t, "", value, "returned non-empty value for expired key")
}

func TestEviction(t *testing.T) {
    c := cache.New(capacity, globalExpiry)

    c.Set("evicted", "value")
    for i := 0; i < capacity; i++ {
        key := fmt.Sprintf("%s%d", "key", i)
        c.Set(key, "value")
    }
    for i := 0; i < capacity; i++ {
        key := fmt.Sprintf("%s%d", "key", i)
        c.Get(key)
    }

    value, ok := c.Get("evicted")
    assert.Equal(t, false, ok, "returned found for expired key")
    assert.Equal(t, "", value, "returned non-empty value for expired key")
}