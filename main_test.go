package main

import (
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "testing"
    "time"

    "github.com/go-redis/redis"
    "github.com/stretchr/testify/assert"
    "github.com/helloworld/redis-proxy/proxy"

)

var (
    testRedis = "redis:6379"
    testCapacity = 10
    testExpiry = 100
    testPort = 8080
    testMaxClients = 10
)

func getRedisClient(t *testing.T) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr: testRedis,
    })

    _, err := client.Ping().Result()
    if err != nil {
        t.Fatal(err)
    }

    return client
}

// Make request to specified path and return 
// Copied from: https://github.com/segmentio/testdemo/blob/master/web/web_test.go#L13
func getRequestBody(path string, s *http.Server, t *testing.T) string {
    res := getResponse(path, s, t)

    body, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        t.Fatal(err)
    }

    return string(body)
}

func getResponse(path string, s *http.Server, t *testing.T) *http.Response {
    // Pick port automatically for parallel tests and to avoid conflicts
    l, err := net.Listen("tcp", ":0")
    if err != nil {
        t.Fatal(err)
    }
    defer l.Close()
    go s.Serve(l) 
    
    res, err := http.Get("http://" + l.Addr().String() + path)
    if err != nil {
        t.Fatal(err)
    }

    return res
}

// Test whether proxy correctly retrieves value from redis
// 1. Add (key1, value1) to redis
// 2. Query key1 from proxy
// 3. Check response = value1
func TestRequestKeyFromProxy(t *testing.T) {
    p, _ := proxy.New(testRedis, testCapacity, testExpiry, testMaxClients)
    s := proxy.NewServer(testPort, p)

    // Add key, value to redis
    client := getRedisClient(t)
    client.Set("key1", "value1", 0)

    // Request value from proxy
    response := getRequestBody("/GET/key1", s, t)

    assert.Equal(t, "value1", string(response), "proxy incorrectly retrieved value from redis")
}

// Test whether cache properly caches content
// 1. Add (key2, value2) to redis
// 2. Query key2 from proxy
// 2. Update (key2, value2modified)
// 3. Query key2 from poxy
// 3. Check response = value2, since it should be cached
func TestRequestKeyFromCache(t *testing.T) {
    p, _ := proxy.New(testRedis, testCapacity, testExpiry, testMaxClients)
    s := proxy.NewServer(testPort, p)

    // Add key, value to redis
    client := getRedisClient(t)
    client.Set("key2", "value2", 0)

    // Request value from proxy
    getRequestBody("/GET/key2", s, t)

    // Update key in redis
    client.Set("key2", "value2modified", 0)

    response := getRequestBody("/GET/key2", s, t)
    assert.Equal(t, "value2", string(response), "proxy incorrectly retrieved value from cache")
}

// Test cache expiry
// 1. Add (key2, value2) to redis
// 2. Query key2 from proxy
// 2. Update (key2, value2modified)
// 3. Query key2 from poxy
// 3. Check response = value2, since it should be cached
func TestCacheExpiry(t *testing.T) {
    p, _ := proxy.New(testRedis, testCapacity, testExpiry, testMaxClients)
    s := proxy.NewServer(testPort, p)

    // Add key, value to redis
    client := getRedisClient(t)
    client.Set("key3", "value3", 0)

    // Request value from proxy
    getRequestBody("/GET/key3", s, t)

    // Update key in redis
    client.Set("key3", "value3modified", 0)

    time.Sleep(200 * time.Millisecond)
    response := getRequestBody("/GET/key3", s, t)

    assert.NotEqual(t, "value3", string(response), "cache did not invalidate expired value")
    assert.Equal(t, "value3modified", string(response), "proxy did not query redis for expired value")
}

// Test cache expiry
// 1. Add (key2, value2) to redis
// 2. Query key2 from proxy
// 2. Update (key2, value2modified)
// 3. Query key2 from poxy
// 3. Check response = value2, since it should be cached
func TestCacheEviction(t *testing.T) {
    p, _ := proxy.New(testRedis, testCapacity, testExpiry, testMaxClients)
    s := proxy.NewServer(testPort, p)

    // Add key, value to redis
    client := getRedisClient(t)
    for i := 0; i < testCapacity + 1; i++ {
        key := fmt.Sprintf("key_eviction_test%d", i)
        value := fmt.Sprintf("value_eviction_test%d", i)
        client.Set(key, value, 0)
    }
    

    for i := 0; i < testCapacity + 1; i++ {
        key_path := fmt.Sprintf("/GET/key_eviction_test%d", i)
        getRequestBody(key_path, s, t)
    }

    for i := 0; i < testCapacity + 1; i++ {
        key := fmt.Sprintf("key_eviction_test%d", i)
        value := fmt.Sprintf("value_eviction_test_updated%d", i)
        client.Set(key, value, 0)
    }

    // Request value from proxy\
    evicted_key_path := fmt.Sprintf("/GET/key_eviction_test%d", 0)
    response := getRequestBody(evicted_key_path, s, t)

    assert.NotEqual(t, "value_eviction_test0", string(response), "cache did not remove LRU entry upon reaching testCapacity")
    assert.Equal(t, "value_eviction_test_updated0", string(response), "proxy did not query redis for entry evicted from cache")
}

func TestNonExistentKey(t *testing.T) {
    p, _ := proxy.New(testRedis, testCapacity, testExpiry, testMaxClients)
    s := proxy.NewServer(testPort, p)

    res := getResponse("/GET/nonexistent", s, t)
    assert.Equal(t, 404, res.StatusCode, "server did not return 404 error for nonexistent key")
}