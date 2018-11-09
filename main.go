package main 

import (
    "flag"
    "fmt"
	"os"

    "github.com/helloworld/redis-proxy/proxy"
)


var port = flag.Int("port", 8080, "bind address")
var capacity = flag.Int("capacity", 100, "cache capacity")
var globalExpiry = flag.Int("global-expiry", 60 * 1000, "cache expiration (in milliseconds)")
var maxClients = flag.Int("max-clients", 5, "maximum number of concurrent clients")
var redisAddress = flag.String("redis-address", "", "redis address")

func main() {
    flag.Parse()

	if r := os.Getenv("REDIS_ADDRESS"); r != "" && *redisAddress == "" {
		*redisAddress = r
	}
	if *redisAddress == "" {
		*redisAddress = "redis:6379"
	}

    // Initialize Proxy
    p, err := proxy.New(*redisAddress, *capacity, *globalExpiry, *maxClients)
    if err != nil {
        return
    }

    // Initialize and start server
    server := proxy.NewServer(*port, p)
    server.ListenAndServe()
    fmt.Println("Server running")
}   
