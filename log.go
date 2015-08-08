package uweb

import (
	"log"
)

const (
	LOG_LEVEL_1 = 1
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

// @impl Middleware
func (lg *Logger) Handle(c *Context) int {
	log.Println("[uweb]", c.Req.IP + "-->", c.Req.Method, c.Req.URL.Path)
	
	c.Next()
	
	if lg.level == LOG_LEVEL_1 {
		log.Println("[uweb]", c.Req.IP + "<--", c.Res.Status)
	} else {
		log.Println("[uweb]", c.Req.IP + "<--", c.Res.Status, string(c.Res.Body))
	}
	
	return NEXT_CONTINUE
}
