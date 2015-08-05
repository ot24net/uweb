package uweb

import (
	"log"
)

const (
	LOG_LEVEL_1 = 1
	LOG_LEVEL_2 = 2
)

//
// Default logger middleware
//
func MdLogger(tag string, level int) Middleware {
	return NewLogger(tag, level)
}

//
// Logger print request and response infor
//
type Logger struct {
	tag   string
	level int
}

// Create default logger
func NewLogger(tag string, level int) *Logger {
	return &Logger{
		tag:   tag,
		level: level,
	}
}

// Print start and end time
func (lg *Logger) Handle(c *Context) int {
	log.Println(lg.tag, c.Req.IP+"-->", c.Req.Method, c.Req.URL.Path)
	c.Next()
	if lg.level == LOG_LEVEL_1 {
		log.Println(lg.tag, c.Req.IP+"<--", c.Res.Status)
	} else {
		log.Println(lg.tag, c.Req.IP+"<--", c.Res.Status, string(c.Res.Body))
	}
	return NEXT_CONTINUE
}
