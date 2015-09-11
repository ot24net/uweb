package uweb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
)

//
// Render interface
//
type Render interface {
	// plain text
	Plain(data string) error

	// Render json format
	//
	// data - will json marshal
	// padding - JSONP padding function name
	//
	// about jsonp see:
	// http://www.cnblogs.com/dowinning/archive/2012/04/19/json-jsonp-jquery.html
	Json(data interface{}, padding string) error

	// Render html format
	//
	// name - Template name
	// data - Data for template
	Html(name string, data interface{}) error
}

//
// Create render middleware
//
func MdRender(root, suffix string) Middleware {
	tpl, err := NewTemplate(root, suffix)
	if err != nil {
		panic(err)
	}
	return tpl
}

//
// Default template
//
var (
	tplHelpers = make(map[string]interface{})
)

// Register helper to default tpl instance
func Helper(name string, f interface{}) {
	if _, ok := tplHelpers[name]; ok {
		panic("Template: DUP helper")
	}
	tplHelpers[name] = f
}

//
// Cached template
//
type Template struct {
	tpl *template.Template
}

// Create empty object
func NewTemplate(root, suffix string) (*Template, error) {
	// walk
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		match := true
		if len(suffix) > 0 {
			if filepath.Ext(path) != suffix {
				match = false
			}
		}
		if match {
			files = append(files, path)
			if DEBUG {
				log.Println(LOG_TAG, "Template: parse file ", path)
			}
		}
		return nil
	})

	// parse
	tpl, err := template.ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	// tpl
	return &Template{
		tpl: tpl,
	}, nil
}

// @impl Midelleware
func (t *Template) Handle(c *Context) int {
	c.Render = &tplRender{c, t}
	return NEXT_CONTINUE
}

// Execute template
func (t *Template) Execute(w io.Writer, name string, data interface{}) error {
	return t.tpl.ExecuteTemplate(w, name, data)
}

//
// Impl Render
//
type tplRender struct {
	c   *Context
	tpl *Template
}

// @impl Render.Html
func (r *tplRender) Html(name string, data interface{}) error {
	// exec
	buf := new(bytes.Buffer)
	r.tpl.Execute(buf, name, data)

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

// @impl Render.Plain
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

// @impl Render.Json
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
