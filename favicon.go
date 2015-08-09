package uweb

import (
	"io/ioutil"
)

//
// Create Favicon middleware
//
func MdFavicon(p string) Middleware {
	if len(p) == 0 {
		panic("len(p) == 0")
	}
	f, err := NewFavicon(p)
	if err != nil {
		panic(err)
	}
	return f
}

//
// Cache favicon data.
// should restart if file changed
//
type Favicon struct {
	icon []byte
}

// Create favicon
func NewFavicon(p string) (*Favicon, error) {
	icon, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return &Favicon{
		icon: icon,
	}, nil
}

// Reponse favicon data if is such req
// @impl Middleware
func (f *Favicon) Handle(c *Context) int {
	// fast check path
	if c.Req.URL.Path != "/favicon.ico" {
		return NEXT_CONTINUE
	}

	// check method
	if c.Req.Method != "GET" && c.Req.Method != "HEAD" {
		if c.Req.Method == "OPTIONS" {
			c.Res.Status = 200
		} else {
			c.Res.Status = 405
		}
		c.Res.Header().Set("Allow", "GET, HEAD, OPTIONS")
		return NEXT_BREAK
	}

	// set header and body
	h := c.Res.Header()
	h.Set("Cache-Control", "public, max-age=86400")
	h.Set("Content-Type", "image/x-icon")

	c.Res.Status = 200
	c.Res.Body = f.icon // will not change underline byte array

	// no need continue
	return NEXT_BREAK
}
