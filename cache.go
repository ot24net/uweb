package uweb

import (
	"time"
	"errors"
	"strings"
	
	"github.com/garyburd/redigo/redis"
)

//
// Cache interface
//
type Cache interface {
	Set(key string, data []byte, expire int) error
	Get(key string) ([]byte, error)
}

// 
// Cache middleware
//
func MdCache(driver, dsn string) Middleware {
	if driver != "redis" {
		panic("cache: only support redis")
	}
	r, err := NewRedisCache(dsn)
	if err != nil {
		panic(err)
	}
	return r
}

//
// RedisCache
//
type RedisCache struct {
	pool *redis.Pool
}

func NewRedisCache(dsn string) (*RedisCache, error) {
	// parse dsn
	arr := strings.Split(dsn, "@")
	if len(arr) != 2 {
		return nil, errors.New("Cache: invalid dsn")
	}
	pwd, addr := arr[0], arr[1]
	
	// pool
	pool := &redis.Pool{
		MaxIdle:     6,
 		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			if len(pwd) > 0 {
				if _, err := c.Do("AUTH", pwd); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	
	// cache
	return &RedisCache{
		pool:   pool,
	}, nil
}

// @impl Middleware
func (r *RedisCache) Handle(c *Context) int {
	c.Cache = r
	return NEXT_CONTINUE
}

// @impl Cache.Set
func (r *RedisCache) Set(key string, data []byte, expire int) error {
	// c
	c := r.pool.Get()
	defer c.Close()
	
	// set
	if _, err := c.Do("SETEX", key, expire, data); err != nil {
		return err
	}
	
	// ok
	return nil
}

// @impl Cache.Get
func (r *RedisCache) Get(key string) ([]byte, error) {
	// c
	c := r.pool.Get()
	defer c.Close()
	
	// get
	value, err := c.Do("GET", key)
	if err != nil {
		return nil, err
	}
	
	// should not return nil data if not err	
	if value == nil {
		return make([]byte, 0), nil
	}
	return value.([]byte), nil
}

