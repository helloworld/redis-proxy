package redis_test

import (
    "fmt"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/helloworld/redis-proxy/redis"
)

var (
    addr = "redis:6379"
)

func seed(r *redis.RedisStore) {
    for i := 0; i < 10; i++ {
        key := fmt.Sprintf("%s%d", "key", i)
        value := fmt.Sprintf("%s%d", "value", i)
        r.Conn.Set(key, value, 0)
    }
}

func TestNewConnection(t *testing.T) {
    _, err := redis.New(addr)
    if err != nil {
        t.Fail()
    }
}

func TestGetNonExistentKey(t *testing.T) {
    r, _ := redis.New(addr)
    seed(r)

    _, err := r.Get("nonexistent")
    if _, ok := err.(*redis.NotFoundError); !ok {
        t.Fail()
    }
}

func TestGetKey(t *testing.T) {
    r, _ := redis.New(addr)
    seed(r)

    value, err := r.Get("key1")
    if err != nil {
        t.Fail()
    }
    assert.Equal(t, "value1", value, "incorrect value retrieved from redis")
}