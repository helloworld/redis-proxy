package cache

import (
    // "fmt"
    "time"
    "sync"
    
    "github.com/golang/groupcache/lru"
)

type CacheValue struct {
    Value   string
    Expiry  int64
}

type CacheStore struct {
    Cache         *lru.Cache
    GlobalExpiry  int64 // Duration in milliseconds after which entry is expired
    mutex         *sync.Mutex
}

func New(capacity int, globalExpiry int) *CacheStore {
    return &CacheStore{
        Cache:        lru.New(capacity),
        GlobalExpiry: int64(globalExpiry),
        mutex:        &sync.Mutex{},
    }
}

// Get retrieves the value for the specified key from the cache. 
// if the entry has not expired.  If the entry is expired, it is
// removed from the cache.
func (s *CacheStore) Get(key string) (string, bool) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if cv, ok := s.Cache.Get(key); ok {
        cacheValue := cv.(*CacheValue)
        if s.IsExpired(cacheValue) {
            s.Remove(key)
            return "", false
        }

        return cacheValue.Value, true
    }

    return "", false
}

// Set adds a given entry to the cache and sets the expiry time
// for the entry to be GlobalExpiry milliseconds from the current time.
func (s *CacheStore) Set(key string, value string) {
    now := time.Now()
    duration := time.Millisecond * time.Duration(s.GlobalExpiry)
    expiry := int64(now.Add(duration).UnixNano())

    cv := &CacheValue{
        Value:  value,
        Expiry: expiry,
    }

    s.mutex.Lock()
    s.Cache.Add(key, cv)
    s.mutex.Unlock()
}

func (s *CacheStore) Remove(key string) {
    s.Cache.Remove(key) 
}

// IsExpired checks whether the current time is past the expiry 
// time set on a CacheValue
func (s *CacheStore) IsExpired(cv *CacheValue) bool {
    now := int64(time.Now().UnixNano()) 
    if cv.Expiry - now >= 0 {
        return false
    }

    return true
}
