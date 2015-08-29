package uweb

import (
	"fmt"
	"log"
	"net/http/httputil"
	"time"
)

const (
	// not print reponse body
	LOG_LEVEL_1 = 1

	// will print reponse body
	LOG_LEVEL_2 = 2
)

//
// Create log middleware
//
func MdLogger(level int) Middleware {
	return NewLogger(level)
}

//
// Logger print request and response
//
type Logger struct {
	level int
}

func NewLogger(level int) *Logger {
	return &Logger{
		level: level,
	}
}

const (
	uweb_log_tag = "[uweb]"
)

// @impl Middleware
func (lg *Logger) Handle(c *Context) int {
	reqBody := "\n"
	if lg.level == LOG_LEVEL_2 {
		dump, err := httputil.DumpRequest(c.Req.Request, true)
		if err != nil {
			panic(err)
		}
		reqBody = fmt.Sprintf("\n{\n\n%s}\n", string(dump))
	}

	log.Printf("%s %s%s %s %s %s", uweb_log_tag, c.Req.IP, "-->", c.Req.Method, c.Req.URL.Path, reqBody)

	start := time.Now()
	c.Next()
	stop := time.Now()

	spend := int64(stop.Sub(start) / time.Millisecond)
	size := len(c.Res.Body)
	resBody := "\n"
	if lg.level == LOG_LEVEL_2 {
		dump := "c.Res.Body == null"
		if size > 0 {
			dump = string(c.Res.Body)
		}
		resBody = fmt.Sprintf("\n{\n\n%s\n\n}\n", dump)
	}
	log.Printf("%s %s%s %s %s %d %d(byte) %d(ms) %s", uweb_log_tag, c.Req.IP, "<--", c.Req.Method, c.Req.URL.Path, c.Res.Status, size, spend, resBody)

	return NEXT_CONTINUE
}
