# redis-proxy


A simple redis proxy, providing a HTTP API and caching layer for redis GET commands. 


Features:
- HTTP API - make redis GET command calls through a HTTP GET request
- Cached GET - GET command requests are cached in an in-memory LRU cache

## Usage

**Prerequisites:** `make`, `docker`, `docker-compose`

**Configuration:** 

Configuration options can be set in the .env file

```
CAPACITY=1000
GLOBAL_EXPIRY=60000
PORT=8080
REDIS_ADDRESS=redis:6379
```

**Running:**

```bash
# clone the repo
git clone git@github.com:helloworld/redis-proxy.git

# cd into repo
cd redis-proxy

# build and run tests
make test

# run in docker container
make run

# stop container
make stop
```
 
**API:**:

- `GET /` shows usage and displays configuration settings for proxy instance
- `GET /GET/{key1}` returns the value associated with `key1`. Returns from cache if available, otherwise retrieves from redis


## Architecture Overview

**Components**

There are three components to redis-proxy:

**1. cache**

The cache is implemented as an LRU cache. It can be easily swapped out There are two available configuration options:

- `CAPACITY`: The maximum number of keys the cache retains. Once the cache fills to capacity, the least recently read key is evicted each time a new key needs to be added to the cache. 

- `REDIS_ADDRESS`: entries in the cache expire if they remain in the cache for a duration longer than the set global-expiry. If an entry is expired, a `HTTP GET` request to the proxy will process the request as if the entry was never stored in the cache. This option is configured in milliseconds. 
  
**2. redis**

The backing redis instance for the proxy is configurable via the following configuration option:

- `REDIS_ADDRESS`: a string with the format "host:port". For example, `localhost:6379`.

**3. HTTP server**

The HTTP server handling `GET` requests. A request to the endpoint `GET /GET/{key1}` first attempts to retrieve a value from the cache, and if unavailable, retrieves the value from redis. 


**Parallel concurrent access:** 

I used a naive implementation to limit the number of clients that are able to concurrently connect. 

There is a semaphore `sema chan struct{}` which is initialized to a size of `maxClients`. On each request, a `struct` is passed to the channel. If the channel is filled to capacity `maxClients`, this will block and execute when there is an available space. 

A limitation is that the clients will queue without bound, until it hits the system limit. With more time, a better concurrent access method can be implemented.


## Code Overview

The code for redis proxy is split into four main packages:

**`cache/cache.go`**

The cache is implemented as an LRU cache. It uses the in-memory LRU cache implementation provided in Google's [`groupcache` library](https://github.com/golang/groupcache). It is not safe for concurrent access, so all accesses must hold a mutex. The `CacheStore` struct stores the pointer to the LRU cache and the mutex. 

The entry expiry time is stored as the unix timestamp at which it becomes invalid. If a `Get(key)` call is made for an expired `key`, it is removed from the cache. The underlying cache implementation handles evicting LRU items. 

**`redis/redis.go`**

This is a lightweight layer over the [`go-redis`](https://github.com/go-redis/redis) redis client. This client handles connection pooling to enable concurrent access without having to worry about it ourselves. 

**`proxy/proxy.go`**

Provides two methods that instantiate a new proxy service. 

- `New(redisAddress string, capacity int, globalExpiry int) (*Proxy, error)
`
Returns a new instance of `Proxy`, which contains an instance of a redis   connection and cache. 

- `NewServer(port int, p *Proxy) *http.Server`
Returns an HTTP server with the two available API endpoints.

**`main.go`** 

Instantiates a new `proxy` and `http.Server` using the configurations options passed through the command line, and starts the server.

## Algorithmic Complexity

**Cache Operations**

The `groupcache/lru` library provides us with `O(1)` amortized lookup and set value. The additional check whether an entry is expired is a `O(1)` operation, making the overall complexity of all cache operations `O(1)`.

**Proxy Operations**

If the requested `key` is available in the cache, we are able to retrieve the value in `O(1)` time.

Otherwise, we must make a request to redis, which provides `O(1)` amortized lookup time if all the data fits in memory, or `O(1+n/k)` where n is the number of items and k the number of buckets [(source)](https://stackoverflow.com/questions/15216897/how-does-redis-claim-o1-time-for-key-lookup).

## Time Spent

**MVP:**  
Setting up Docker - 15 minutes  
Implementing cache - 30 minutes  
Implementing server - 30 minutes  

**Final version:**  
Writing cache.go and redis.go - 30 minutes  
Writing proxy.go - 1 hour  
Testing - 1 hour  
Debugging (total) - 1 hour  
Documentation - 30 minutes  

## Omitted requirements:

Unfortunately, I did not have the time to implement the `Redis client protocol` requirement.