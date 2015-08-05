package uweb

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
)

//
// Redis-based session middleware
//
func MdSession(addr, pwd string, expire int) Middleware {
	rs, err := NewRedisStore(addr, pwd, expire)
	if err != nil {
		panic(err)
	}
	return rs
}

//
// Session Cache
//
type SesCache map[string]string

// marshal for store
func (sc SesCache) Marshal() ([]byte, error) {
	return json.Marshal(sc)
}

// unmarshal from store
func (sc *SesCache) Unmarshal(from []byte) error {
	return json.Unmarshal(from, sc)
}

//
// Session store interface
//
type SesStore interface {
	Load(string, *SesCache) error
	Save(string, *SesCache) error
}

//
// Session is per request sesssion
//
type Session struct {
	sid   string
	cache SesCache
	dirty bool
	store SesStore
}

// Create new session
func NewSession(sid string, store SesStore) (*Session, error) {
	s := &Session{
		sid:   sid,
		cache: make(SesCache),
		store: store,
	}
	if len(s.sid) == 0 {
		s.sid = genSid()
	} else {
		if err := s.store.Load(s.sid, &s.cache); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// create random sid
func genSid() string {
	k := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return ""
	} else {
		return base64.StdEncoding.EncodeToString(k)
	}
}

// Get sid
func (s *Session) Id() string {
	return s.sid
}

// Set item
func (s *Session) Set(k, v string) {
	s.cache[k] = v
	s.dirty = true
}

// Get item
func (s *Session) Get(k string) string {
	v, _ := s.cache[k]
	return v
}

// Del item
func (s *Session) Del(k string) {
	delete(s.cache, k)
	s.dirty = true
}

// Save all
func (s *Session) Save() error {
	if !s.dirty {
		return nil
	}
	if err := s.store.Save(s.sid, &s.cache); err != nil {
		return err
	}
	s.dirty = false
	return nil
}

//
// RedisStore todo: use redis, not mem
//
type RedisStore struct {
	addr   string
	pwd    string
	expire int // expire seconds
	pool   *redis.Pool
}

// Create redis store
func NewRedisStore(addr, pwd string, expire int) (*RedisStore, error) {
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

	return &RedisStore{
		addr:   addr,
		pwd:    pwd,
		expire: expire,
		pool:   pool,
	}, nil
}

const (
	SID_COOKIE_KEY = "_sid"
)

// Impl Middleware
func (rs *RedisStore) Handle(c *Context) int {
	// sid
	sid := ""
	k, err := c.Req.Cookie(SID_COOKIE_KEY)
	if err == nil && k != nil {
		sid = k.Value
	}

	// ses
	sess, err := NewSession(sid, rs)
	if err != nil {
		c.Res.Status = http.StatusInternalServerError
		c.Res.Err = err
		return NEXT_BREAK
	}

	// ok
	c.Sess = sess
	if sid == "" { // new sid
		c.Res.SetCookie(SID_COOKIE_KEY, sess.Id())
	}
	return NEXT_CONTINUE
}

// Load from redis
func (rs *RedisStore) Load(sid string, val *SesCache) error {
	// c
	c := rs.pool.Get()
	defer c.Close()

	// get
	data, err := c.Do("GET", sid)
	if err != nil {
		return err
	}

	// unmarshal
	if data != nil {
		src := data.([]byte)
		if err := val.Unmarshal(src); err != nil {
			return err
		}
	}

	// ok
	return nil
}

// Save to redis
func (rs *RedisStore) Save(sid string, val *SesCache) error {
	// marshal
	data, err := val.Marshal()
	if err != nil {
		return err
	}

	// c
	c := rs.pool.Get()
	defer c.Close()

	// save
	if _, err := c.Do("SETEX", sid, rs.expire, data); err != nil {
		return err
	}
	return nil
}
