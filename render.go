package uweb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
	"sync"
)

//
// Template data map
//
type TplMap map[string]interface{}

//
// Render interface
//
type Render interface {
	// plain text
	Plain(data string) error

	// json or jsonp if padding not empty, about jsonp see:
	// http://www.cnblogs.com/dowinning/archive/2012/04/19/json-jsonp-jquery.html
	Json(data interface{}, padding string) error

	// Usage:
	//  c.Render.Html("home", uweb.TplMap{
	//    "common/header.html": data.header,
	//    "home/content.html": data.content,
	//    "common/footer.html": data.footer,
	//  })
	Html(name string, data TplMap) error
}

//
// Create render middleware
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
		panic("Template: DUP helper")
	}
	t.helpers[name] = f
}

// Parse files and register helper funcs
func (t *Template) Parse(name string, data TplMap) (*template.Template, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// in cache?
	if tpl, ok := t.cache[name]; ok {
		return tpl, nil
	}

	// parse files
	var files []string
	for k, _ := range data {
		files = append(files, filepath.Join(t.dir, k))
	}
	tpl, err := template.ParseFiles(files...)
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
func (r *tplRender) Html(name string, data TplMap) error {
	// verify
	if len(data) == 0 {
		panic("empty data is not allowed")
	}

	// tpl
	tpl, err := r.tpl.Parse(name, data)
	if err != nil {
		return err
	}

	// buf
	buf := new(bytes.Buffer)
	for k, v := range data {
		tpl.ExecuteTemplate(buf, k, v)
	}

	// w
	w := r.c.Res

	// set body header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//  body
	w.Body = buf.Bytes()

	// status
	if w.Status == 0 {
		w.Status = 200
	}

	// ok
	return nil
}

// Plain text
func (r *tplRender) Plain(data string) error {
	// w
	w := r.c.Res

	// body
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Body = []byte(data)

	// status
	if w.Status == 0 {
		w.Status = 200
	}

	// ok
	return nil
}

// Render json
func (r *tplRender) Json(v interface{}, padding string) error {
	// w
	w := r.c.Res

	// body
	result, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if len(padding) > 0 {
		result = []byte(fmt.Sprintf("%s(%s);", padding, string(result)))
	}
	w.Body = result

	// header
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// status
	if w.Status == 0 {
		w.Status = 200
	}

	// ok
	return nil
}
