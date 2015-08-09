package uweb

import (
	"net/url"
	"path"
	"strings"
)

//
// Create redirector middleware
//
func MdRedirect() Middleware {
	return new(Redirector)
}

//
// Redirect to location
//
type Redirector struct {
	// empy
}

// @impl Middleware
func (r *Redirector) Handle(c *Context) int {
	c.Redirect = &Redirect{c}
	return NEXT_CONTINUE
}

//
// Redirect closure
//
type Redirect struct {
	c *Context
}

// Redirect to url
func (r *Redirect) To(urlStr string) {
	// req, res
	req, res := r.c.Req, r.c.Res

	// copy from http/server.go
	// Location should be an absolute URI, like
	if u, err := url.Parse(urlStr); err == nil {
		oldpath := req.URL.Path
		if oldpath == "" { // should not happen, but avoid a crash if it does
			oldpath = "/"
		}
		if u.Scheme == "" {
			// no leading http://server
			if urlStr == "" || urlStr[0] != '/' {
				// make relative path absolute
				olddir, _ := path.Split(oldpath)
				urlStr = olddir + urlStr
			}
			var query string
			if i := strings.Index(urlStr, "?"); i != -1 {
				urlStr, query = urlStr[:i], urlStr[i:]
			}
			// clean up but preserve trailing slash
			trailing := strings.HasSuffix(urlStr, "/")
			urlStr = path.Clean(urlStr)
			if trailing && !strings.HasSuffix(urlStr, "/") {
				urlStr += "/"
			}
			urlStr += query
		}
	}

	// RFC2616 recommends that a short note "SHOULD" be included in the
	// response because older user agents may not understand 301/307.
	// Shouldn't send the response for POST or HEAD; that leaves GET.
	res.Status = 302
	res.Header().Set("Location", urlStr)
	if req.Method == "GET" {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.Body = []byte("Redirecting to " + urlStr + ".")
	}
}

// Back to referrer or "/"
func (r *Redirect) Back() {
	urlStr := r.c.Req.Header.Get("Referrer")
	if len(urlStr) == 0 {
		urlStr = "/"
	}
	r.To(urlStr)
}
