package redis

import (
    "fmt"

    "github.com/go-redis/redis"
)

type RedisStore struct {
    Conn *redis.Client
}


type NotFoundError struct {
    Key string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("key not found %s", e.Key)
}


func New(addr string) (*RedisStore, error) {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
    })

    _, err := client.Ping().Result()
    if err != nil {
        return nil, err
    }

    return &RedisStore{
        Conn: client,
    }, nil
}

func (r *RedisStore) Get(key string) (string, error) {
    value, err := r.Conn.Get(key).Result()

    if err != nil {
        if err == redis.Nil {
            return "", &NotFoundError{key}
        }

        return "", err
    }

    return value, nil
}

