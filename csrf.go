package uweb

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

const (
	// secret in session
	CSRF_SECRET_KEY = "_csrf_secret"

	// token in session
	CSRF_TOKEN_KEY = "_csrf_token"
)

const (
	// the longer the better
	CSRF_SECRET_LEN = 18

	// doesn't need to be long
	CSRF_SALT_LEN = 8
)

//
// CSRF middleware, depends on session
//
func MdCsrf() Middleware {
	return NewCsrf()
}

//
// CSRF protect
//
type Csrf struct {
	// empty
}

// Create csrf handler
func NewCsrf() *Csrf {
	return new(Csrf)
}

// Impl Middleware
func (cf *Csrf) Handle(c *Context) int {
	// lazily creates a csrf token
	// create one per session
	secret := c.Sess.Get(CSRF_SECRET_KEY)
	if len(secret) == 0 {
		// create new token
		secret = cf.genSecret(CSRF_SECRET_LEN)
		salt := cf.genSalt(CSRF_SALT_LEN)
		token := cf.genToken(salt, secret)

		// save in session
		c.Sess.Set(CSRF_SECRET_KEY, secret)
		c.Sess.Set(CSRF_TOKEN_KEY, token)

		// for angular.js
		c.Res.SetCookie("XSRF-TOKEN", token)
	}

	// ignore method
	switch c.Req.Method {
	case "GET", "HEAD", "OPTIONS":
		return NEXT_CONTINUE
	}

	// parse token
	token := c.Req.FormValue("_csrf")
	if len(token) == 0 {
		h := c.Req.Header
		token = h.Get("X-CSRF-Token")
		if len(token) == 0 {
			token = h.Get("X-XSRF-Token")
		}
	}
	if len(token) == 0 {
		c.Res.Status = 400
		c.Res.Err = errors.New("Csrf: no csrf")
		return NEXT_BREAK
	}

	// verify
	if err := cf.verify(secret, token); err != nil {
		c.Res.Status = 403
		c.Res.Err = err
		return NEXT_BREAK
	}

	// ok
	return NEXT_CONTINUE
}

// create a secret key
// this __should__ be cryptographically secure,
// but generally client's can't/shouldn't-be-able-to access this so it really doesn't matt
func (cf *Csrf) genSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

// create a random salt
func (cf *Csrf) genSalt(length int) string {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

// create a csrf token
func (cf *Csrf) genToken(salt, secret string) string {
	h := sha1.New()
	io.WriteString(h, salt)
	io.WriteString(h, "-")
	io.WriteString(h, secret)
	return salt + "-" + base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (cf *Csrf) verify(secret, token string) error {
	// extract salt
	a := strings.SplitN(token, "-", 2)
	if len(a) != 2 {
		return errors.New("Csrf: invalid token")
	}
	salt := a[0]
	if len(salt) == 0 {
		return errors.New("Csrf: empty salt")
	}

	// token
	expected := cf.genToken(salt, secret)
	if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
		return errors.New("Csrf: invalid token")
	}

	// ok
	return nil
}
