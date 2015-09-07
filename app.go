package uweb

import (
	"log"
	"net/http"
	"sync"
)

// static const values
const (
	VERSION = "0.9.1"
)

// global ctrl values
var (
	DEBUG = true
)

//
// Web application
// store global objects, such as middleware
//
type Application struct {
	mws  []Middleware // all middlewares
	pool sync.Pool    // cache Context
}

// Create empty application without any middleware
func NewApp() *Application {
	// app
	app := &Application{
		mws: make([]Middleware, 0),
	}
	// pool
	app.pool.New = func() interface{} {
		return NewContext(app)
	}
	// ok
	return app
}

// Add one middleware
func (a *Application) Use(m Middleware) {
	a.mws = append(a.mws, m)
}

// Listen and start serve
func (a *Application) Listen(addr string) error {
	if DEBUG {
		log.Println(LOG_TAG, "Application: listen at", addr)
	}
	return http.ListenAndServe(addr, a)
}

// Handle all http request
// @impl http.Handler
func (a *Application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// get c
	c := a.pool.Get().(*Context)

	// run all middlewares and end the response
	c.Req = NewRequest(req)
	c.Res = NewResponse(w)
	if c.Next() != NEXT_ABORT {
		c.Res.End(c.Req)
	}

	// put c, do not forget reset before put
	c.Reset()
	a.pool.Put(c)
}
