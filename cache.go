package uweb

import (
	"errors"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	//"github.com/garyburd/redigo/redis"
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
	switch driver {
	case "memcache":
		r, err := NewMemCache(dsn)
		if err != nil {
			panic(err)
		}
		return r
		/*
			case "redis":
				r, err := NewRedisCache(dsn)
				if err != nil {
					panic(err)
				}
				return r
		*/
	}
	panic("unknow driver")
	return nil
}

//
// MemCache
//
type MemCache struct {
	mc *memcache.Client
}

func NewMemCache(dsn string) (*MemCache, error) {
	return &MemCache{
		mc: memcache.New(dsn),
	}, nil
}

// @impl Middleware
func (m *MemCache) Handle(c *Context) int {
	c.Cache = m
	return NEXT_CONTINUE
}

// @impl Cache.Set
func (m *MemCache) Set(key string, data []byte, expire int) error {
	return m.mc.Set(&memcache.Item{Key: key, Value: data, Expiration: int32(expire)})
}

// @impl Cache.Get
func (m *MemCache) Get(key string) ([]byte, error) {
	item, err := m.mc.Get(key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}

/*

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
		pool: pool,
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

*/
