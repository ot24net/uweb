package uweb

//
// error pages
//
func MdErrPage(data Map) Middleware {
	return &errPage{
		data: data,
	}
}

//
// support 404
//
type errPage struct {
	data Map
}

func (e *errPage) Name() string {
	return "errors"
}

func (e *errPage) Handle(c *Context) int {
	if c.Req.Method != "GET" {
		return NEXT_CONTINUE
	}
	c.Next()

	switch c.Res.Status {
	case 404:
		c.Res.Err = nil
		c.Render.Html(404, "errors/404", e.data)
	}

	return NEXT_CONTINUE
}
