package uweb

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
)

var (
	// cookie id for session
	SID_COOKIE_KEY = "_sid"
)

//
// Session depends on Cache
//
func MdSession(expire int) Middleware {
	m, err := NewSessMan(expire)
	if err != nil {
		panic(err)
	}
	return m
}

//
// Session manager
//
type SessMan struct {
	expire int
}

// Create session manger instance
func NewSessMan(expire int) (*SessMan, error) {
	return &SessMan{
		expire: expire,
	}, nil
}

// @impl Middleware
func (m *SessMan) Handle(c *Context) int {
	// read sid from cookie
	sid, newSess := "", true
	if k, err := c.Req.Cookie(SID_COOKIE_KEY); err == nil && k != nil {
		sid = k.Value
	}
	if len(sid) > 0 {
		newSess = false
	}

	// session
	s := NewSession(sid)
	if newSess {
		c.Res.SetCookie(SID_COOKIE_KEY, s.Id())
	} else {
		if err := s.restore(c.Cache); err != nil {
			// if memcache not start, and sid exist in cookie,
			// make it as new session
			if err != ErrCacheMiss {
				c.Res.Status = 500
				c.Res.Err = err
				return NEXT_BREAK
			}
		}
	}
	c.Sess = s

	// next
	c.Next()

	// save session
	if err := s.save(c.Cache, m.expire); err != nil {
		c.Res.Status = 500
		c.Res.Err = err
		return NEXT_BREAK
	}

	// ok
	return NEXT_CONTINUE
}

//
// Session is per request sesssion
//
type Session struct {
	sid   string
	data  map[string]string
	dirty bool
}

// Create new session
func NewSession(sid string) *Session {
	if len(sid) == 0 {
		sid = genSid()
	}
	s := &Session{
		sid:  sid,
		data: make(map[string]string),
	}
	return s
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
	s.data[k] = v
	s.dirty = true
}

// Get item
func (s *Session) Get(k string) string {
	v, ok := s.data[k]
	if !ok {
		return ""
	}
	return v
}

// Del item
func (s *Session) Del(k string) {
	if _, ok := s.data[k]; ok {
		delete(s.data, k)
		s.dirty = true
	}
}

// Restore from cache
func (s *Session) restore(cache Cache) error {
	data, err := cache.Get(s.sid)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &s.data); err != nil {
		return err
	}
	return nil
}

// Save to cache
func (s *Session) save(cache Cache, expire int) error {
	if !s.dirty {
		return nil
	}
	data, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	return cache.Set(s.sid, data, expire)
}
