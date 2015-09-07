package uweb

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

//
// Params
//
type Params map[string]string

// Convert to int value, if fail return 0
func (p Params) Int(key string) int {
	s, ok := p[key]
	if !ok {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Println(LOG_TAG, "Params: Int err", err)
		return 0
	}
	return v
}

// Get string value, if fail return ""
func (p Params) Str(key string) string {
	s, ok := p[key]
	if !ok {
		return ""
	}
	return s
}

//
// Wrap http request
//
type Request struct {
	// embbed request for convenient
	*http.Request

	// client ip
	IP string

	// url pattern params, Router middleware will set it
	Params Params
}

// Create request
func NewRequest(req *http.Request) *Request {
	return &Request{req, readIp(req), nil}
}

// parse real ip if possible
// may use nginx to set header fields
func readIp(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		i := strings.Index(v, ", ")
		if i == -1 {
			i = len(v)
		}
		return v[:i]
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	return ""
}
