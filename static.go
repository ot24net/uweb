package uweb

import (
	"net/http"
	"path/filepath"
	"strings"
)

//
// Static file server middleware
//
func MdStatic(path, dir string) Middleware {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	return NewStatic(path, dir)
}

//
// Static file server, only suite for small project
// If your web site is busy, just use CDN.
//
// TODO: support compress
//
type Static struct {
	path string // path prefix for statics
	dir  string // static files abs path
}

// Create file server
func NewStatic(path, dir string) *Static {
	return &Static{
		path: path,
		dir:  dir,
	}
}

// @impl Middleware
func (s *Static) Handle(c *Context) int {
	p := c.Req.URL.Path
	if strings.HasPrefix(p, s.path) {
		if len(p) <= len(s.path) {
			return NEXT_CONTINUE
		}
		file := filepath.Join(s.dir, p[len(s.path):])
		http.ServeFile(c.Res, c.Req.Request, file)
		return NEXT_ABORT
	}
	return NEXT_CONTINUE
}
