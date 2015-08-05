package uweb

import (
	"bytes"
	"encoding/json"
	"html/template"
	"path/filepath"
	"sync"
)

//
// Render interface
//
type Render interface {
	Plain(data string) error
	Json(data interface{}) error
	Html(file string, data interface{}) error
}

//
// Render middleware
//
func MdRender(dir string) Middleware {
	defaultTpl.BaseBy(dir)
	return defaultTpl
}

//
// Default template
//
var (
	defaultTpl = NewTemplate()
)

// Register helper to default tpl instance
func Helper(name string, f interface{}) {
	defaultTpl.Helper(name, f)
}

//
// Cached template
//
type Template struct {
	dir     string
	helpers map[string]interface{}

	mu    sync.Mutex // protect cache
	cache map[string]*template.Template
}

// Create empty object
func NewTemplate() *Template {
	return &Template{
		helpers: make(map[string]interface{}),
		cache:   make(map[string]*template.Template),
	}
}

// Set root dir
func (t *Template) BaseBy(dir string) {
	t.dir = dir
}

// Register helper funcs
func (t *Template) Helper(name string, f interface{}) {
	if _, ok := t.helpers[name]; ok {
		panic("dup helper for name")
	}
	t.helpers[name] = f
}

// Parse files and register helper funcs
func (t *Template) Parse(name string) (*template.Template, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// cached
	if tpl, ok := t.cache[name]; ok {
		return tpl, nil
	}

	// new
	tpl, err := template.ParseFiles(filepath.Join(t.dir, name))
	if err != nil {
		return nil, err
	}
	t.cache[name] = tpl

	// register helpers
	if len(t.helpers) > 0 {
		tpl.Funcs(t.helpers)
	}

	// ok
	return tpl, nil
}

// @impl Midelleware
func (t *Template) Handle(c *Context) int {
	c.Render = &tplRender{c, t}
	return NEXT_CONTINUE
}

//
// Impl Render
//
type tplRender struct {
	c   *Context
	tpl *Template
}

// Render html
func (r *tplRender) Html(name string, data interface{}) error {
	// tpl
	tpl, err := r.tpl.Parse(name)
	if err != nil {
		return err
	}

	// execute to buf
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return err
	}

	// set body header
	w := r.c.Res
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Body = buf.Bytes()

	// status
	if r.c.Res.Status == 0 {
		r.c.Res.Status = 200
	}

	// ok
	return nil
}

// Plain text
func (r *tplRender) Plain(data string) error {
	// body
	w := r.c.Res
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Body = []byte(data)

	// status
	if r.c.Res.Status == 0 {
		r.c.Res.Status = 200
	}

	// ok
	return nil
}

// Render json
func (r *tplRender) Json(v interface{}) error {
	// marshal
	result, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// header & body
	w := r.c.Res
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Body = result

	// status
	if r.c.Res.Status == 0 {
		r.c.Res.Status = 200
	}

	// ok
	return nil
}
