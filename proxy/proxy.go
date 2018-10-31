package proxy 

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "github.com/helloworld/redis-proxy/redis"
    "github.com/helloworld/redis-proxy/cache"

)

type Proxy struct {
    db          *redis.RedisStore
    cache       *cache.CacheStore
    maxClients  int
    sema        chan struct{}
}

func New(redisAddress string, capacity int, globalExpiry int, maxClients int) (*Proxy, error) {
   r, err := redis.New(redisAddress)
    if err != nil {
        return nil, err
    }

    c := cache.New(capacity, globalExpiry)

    return &Proxy{
        db: r,
        cache: c,
        maxClients: maxClients,
        sema: make(chan struct{}, maxClients),
    }, nil
}

func NewServer(port int, p *Proxy) *http.Server{
    router := mux.NewRouter()
    router.HandleFunc("/", p.IndexHandler).Methods("GET")
    router.HandleFunc("/GET/{key}", p.GetHandler)

    server := &http.Server{
        Addr: ":" + strconv.Itoa(port),
        Handler: router,
    }

    return server
}

func (prox *Proxy) IndexHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "redis-proxy")
    fmt.Fprintln(w, "usage:")
    fmt.Fprintln(w, "\t GET /GET/{key} - retrieves key from redis")
    fmt.Fprintln(w, "configuration:")
    fmt.Fprintf(w, "\tcapacity: %d\n", prox.cache.Cache.MaxEntries)
    fmt.Fprintf(w, "\tcache expiry: %d ms", prox.cache.GlobalExpiry)
}

func (prox *Proxy) GetHandler(w http.ResponseWriter, r *http.Request) {
    prox.sema <- struct{}{}
    defer func() { <-prox.sema }()
    
    defer r.Body.Close()

    vars := mux.Vars(r)
    key, _ := vars["key"]


    // Retrieve value from cache, if it exists
    if value, ok := prox.cache.Get(key); ok {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(value))
        return
    }

    // Retrieve value from redis
    value, err := prox.db.Get(key)
    if err != nil {
        if keyError, ok := err.(*redis.NotFoundError); ok {
            w.WriteHeader(http.StatusNotFound)
            w.Write([]byte("404: Requested key not found: " + keyError.Key))
            return 
        }
        
        fmt.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("500: Internal Service Error"))
        return   
    }

    prox.cache.Set(key, value)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(value))
    return    
}
