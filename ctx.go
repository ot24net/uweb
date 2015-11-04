package uweb

import (
	"log"
)

//
// Per request context
//
type Context struct {
	// middleware
	app    *Application
	cursor int

	// req & res
	Req *Request
	Res *Response

	// sess & flash
	Cache Cache
	Sess  *Session
	Flash *Flash

	// view
	Locale *Locale
	Render Render

	// route
	Redirect *Redirect
}

// Create empty context, need middleware to
// fullfill its fields
func NewContext(app *Application) *Context {
	return &Context{
		app:    app,
		cursor: -1, // not 0
	}
}

// Reset fields for recycle and reuse
func (c *Context) Reset() {
	c.cursor = -1

	c.Req = nil
	c.Res = nil

	c.Sess = nil
	c.Flash = nil

	c.Locale = nil
	c.Redirect = nil
}

// Next run next middlewares or break out all if
// one return false
func (c *Context) Next() int {
	ret := NEXT_BREAK
	s := len(c.app.mws)
	for {
		c.cursor++
		if c.cursor >= s {
			break
		}
		ret = c.app.mws[c.cursor].Handle(c)
		if ret != NEXT_CONTINUE {
			log.Println(LOG_TAG, "Context: break at cursor", c.cursor)
			c.cursor = s
		}
	}
	return ret
}
