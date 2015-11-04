package uweb

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//
// Render interface
//
type Render interface {
	// Render html format
	//
	// name - Template name
	// data - Data for template
	Html(status int, name string, data interface{}) error
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
	root   string
	suffix string
	tpl    *template.Template
}

// Create empty object
func NewTemplate(root, suffix string) (*Template, error) {
	// tpl
	t := &Template{
		root:   root,
		suffix: suffix,
	}
	if !DEVELOPMENT {
		if err := t.loadTpls(); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *Template) loadTpls() error {
	var files []string
	if err := filepath.Walk(t.root, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		match := true
		if len(t.suffix) > 0 {
			if filepath.Ext(path) != t.suffix {
				match = false
			}
		}
		if match {
			if strings.HasPrefix(filepath.Base(path), ".") {
				match = false
			}
		}

		if match {
			files = append(files, path)
			if DEBUG {
				log.Println(LOG_TAG, "Template: parse file ", path, filepath.Base(path))
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// parse
	tpl, err := template.ParseFiles(files...)
	if err != nil {
		return err
	}

	// ok
	t.tpl = tpl
	return nil
}

// @impl Midelleware
func (t *Template) Handle(c *Context) int {
	c.Render = &tplRender{c, t}
	return NEXT_CONTINUE
}

// Execute template
func (t *Template) Execute(w io.Writer, name string, data interface{}) error {
	if DEVELOPMENT {
		t.loadTpls()
	}
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
func (r *tplRender) Html(status int, name string, data interface{}) error {
	buf := new(bytes.Buffer)
	if err := r.tpl.Execute(buf, name, data); err != nil {
		return err
	}
	r.c.Res.Html(status, buf.Bytes())
	return nil
}

//
// for convinent
//
type Map map[string]interface{}
