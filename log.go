package uweb

import (
	"log"
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
	log.Println(uweb_log_tag, c.Req.IP+"-->", c.Req.Method, c.Req.URL.Path)

	c.Next()

	if lg.level == LOG_LEVEL_1 {
		log.Println(uweb_log_tag, c.Req.IP+"<--", c.Res.Status)
	} else {
		log.Println(uweb_log_tag, c.Req.IP+"<--", c.Res.Status, string(c.Res.Body))
	}

	return NEXT_CONTINUE
}
