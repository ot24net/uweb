package uweb

//
// Create cors middleware
//
// opts params:
//  - origin: Access-Control-Allow-Origin, default is '*'
//  - credentials: Access-Control-Allow-Credentials
//  - maxAge: Access-Control-Max-Age, in seconds
//  - allowMethods: Access-Control-Allow-Methods, default is GET,HEAD,PUT,POST,DELETE
//  - allowHeaders: Access-Control-Allow-Headers
//  - exposeHeaders: Access-Control-Expose-Headers
//
func MdCors(opts map[string]string) Middleware {
	if opts == nil {
		panic("opts == nil")
	}
	return &Cors{
		opts: opts,
	}
}

//
// Default CORS options
//
var DefaultCors = map[string]string{
	"origin":       "*",
	"allowMethods": "GET,HEAD,PUT,POST,DELETE",
}

//
// Cors handler
//
type Cors struct {
	// It's fine if multiple goroutines read from a map simultaneously.
	// But if one goroutine reads from a map while another writes to a map, or if
	// two goroutines write to a map, then the program must synchronize those
	// goroutines in some way.
	opts map[string]string
}

// @impl Middleware
//
// see koa's cors:
// https://github.com/koajs/cors
//
func (co *Cors) Handle(c *Context) int {
	// if the Origin header is not present terminate this set of steps.
	// the request is outside the scope of this specification.
	reqOrigin := c.Req.Header.Get("Origin")
	if len(reqOrigin) == 0 {
		return NEXT_CONTINUE
	}

	// next
	c.Next()

	// h
	h := c.Res.Header()

	// origin
	origin, _ := co.opts["origin"]
	if len(origin) == 0 {
		origin = reqOrigin
	}

	// preflight request
	if c.Req.Method == "OPTIONS" {
		// if there is no Access-Control-Request-Method header or if parsing failed,
		// do not set any additional headers and terminate this set of steps.
		// the request is outside the scope of this specification.
		if len(c.Req.Header.Get("Access-Control-Request-Method")) == 0 {
			return NEXT_CONTINUE
		}

		// origin
		h.Set("Access-Control-Allow-Origin", origin)

		// credentials
		if credentials, ok := co.opts["credentials"]; ok && len(credentials) > 0 {
			h.Set("Access-Control-Allow-Credentials", credentials)
		}

		// maxAge
		if maxAge, ok := co.opts["maxAge"]; ok && len(maxAge) > 0 {
			h.Set("Access-Control-Max-Age", maxAge)
		}

		// allowMethods
		if allowMethods, ok := co.opts["allowMethods"]; ok && len(allowMethods) > 0 {
			h.Set("Access-Control-Allow-Methods", allowMethods)
		} else {
			h.Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,POST,DELETE")
		}

		// allowHeaders
		if allowHeaders, ok := co.opts["allowHeaders"]; ok && len(allowHeaders) > 0 {
			h.Set("Access-Control-Allow-Headers", allowHeaders)
		} else {
			h.Set("Access-Control-Allow-Headers", c.Req.Header.Get("Access-Control-Request-Headers"))
		}

		// other request
	} else {
		// origin
		h.Set("Access-Control-Allow-Origin", origin)

		// credentials
		if credentials, ok := co.opts["credentials"]; ok && len(credentials) > 0 {
			h.Set("Access-Control-Allow-Credentials", credentials)
		}

		// exposeHeaders
		if exposeHeaders, ok := co.opts["exposeHeaders"]; ok && len(exposeHeaders) > 0 {
			h.Set("Access-Control-Expose-Headers", exposeHeaders)
		}
	}

	// ok
	return NEXT_CONTINUE
}
