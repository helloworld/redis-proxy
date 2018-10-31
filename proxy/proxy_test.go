package proxy_test

import (
    "net"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/helloworld/redis-proxy/proxy"
)

var (
    redisAddress = "redis:6379"
    capacity = 10
    globalExpiry = 100
    port = 8080
    maxClients = 10
)

func TestProxyInit(t *testing.T) {
    _, err := proxy.New(redisAddress, capacity, globalExpiry, maxClients)
    if err != nil {
        t.Fatal(err)
    }
}

func TestServer(t *testing.T) {
    p, _ := proxy.New(redisAddress, capacity, globalExpiry, maxClients)
    s := proxy.NewServer(port, p)

    l, err := net.Listen("tcp", ":0")
    if err != nil {
        t.Fatal(err)
    }
    defer l.Close()
    go s.Serve(l) 
    
    res, err := http.Get("http://" + l.Addr().String() + "/")
    if err != nil {
        t.Fatal(err)
    }

    assert.Equal(t, http.StatusOK, res.StatusCode, "server not initialized - non 200 status")
}