package uweb

import (
	"net/http"
	"strings"
)

//
// Wrap http request
//
type Request struct {
	// embbed request for convenient
	*http.Request

	// client ip
	IP string

	// url pattern params, Router middleware will set it
	Params map[string]string
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
