package uweb

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

// if body length less than this, no need to compress
var (
	GZIP_THRESHOLD = 150
)

//
// Compress middleware, only support gzip.
//
func MdCompress() Middleware {
	return NewGzip()
}

//
// @impl(http.ResponseWriter)
//
type gzipWriter struct {
	http.ResponseWriter
	w *gzip.Writer
}

// hide Write in http.ResponseWriter
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.w.Write(data)
}

//
// Gzip compress
//
type Gzip struct {
	pool sync.Pool // cache gzip.Writer
}

// Create gzip middleware
func NewGzip() *Gzip {
	g := new(Gzip)
	g.pool.New = func() interface{} {
		return gzip.NewWriter(nil)
	}
	return g
}

// @impl Middleware
func (g *Gzip) Handle(c *Context) int {
	// bypass some files
	if g.bypass(c.Req) {
		return NEXT_CONTINUE
	}

	// next to get response data
	c.Next()

	// if error
	if c.Res.Err != nil {
		return NEXT_CONTINUE
	}
	// small body
	if len(c.Res.Body) < GZIP_THRESHOLD {
		return NEXT_CONTINUE
	}
	// empty status
	switch c.Res.Status {
	case 204, 205, 304:
		return NEXT_CONTINUE
	}
	// if compressed
	if len(c.Req.Header.Get("Content-Encoding")) > 0 {
		return NEXT_CONTINUE
	}

	// set headers
	h := c.Res.Header()
	h.Set("Vary", "Accept-Encoding")
	h.Set("Content-Encoding", "gzip")
	h.Del("Content-Length")

	// write and close
	rw := c.Res.ResponseWriter
	gw := g.pool.Get().(*gzip.Writer)
	gw.Reset(rw)
	c.Res.ResponseWriter = &gzipWriter{rw, gw}
	c.Res.Close = func() {
		gw.Flush() // flush pending buffer
		g.pool.Put(gw)
	}

	// ok
	return NEXT_CONTINUE
}

// by pass some requests
func (g *Gzip) bypass(req *Request) bool {
	// accept encoding?
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		return true
	}

	// ignore HEAD
	if req.Method == "HEAD" || req.Method == "OPTIONS" {
		return true
	}

	// ignore websocket
	if len(req.Header.Get("Sec-WebSocket-Key")) > 0 {
		return true
	}
	
	// ok
	return false
}
