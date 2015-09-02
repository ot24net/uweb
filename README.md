#uweb Web Framework

uweb is a web framework written in Golang. 
It borrows many ideas from Koa.js, Gin, Playframework, etc.

## example
```
//
// src/main.go
// 
package main

import (
	"github.com/ot24net/uweb"
	_ "ctrls/auth"
)

func main() {
	// app
	app := uweb.NewApp()
	
	// Ignore some path
	app.Use(uweb.MdIgnore([]string{"/haproxy"}))
	
	// Response favicon 
	app.Use(uweb.MdFavicon("../../pub/img/favicon.ico"))
	
	// Serve static files, "/pub" is path prefix, and "../../pub" is file directory
	app.Use(uweb.MdStatic("/pub", "../../pub")) // before compress
	
	// Compress use gzip, currently cannot work with MdSatic	
	app.Use(uweb.MdCompress())
	
	// log
	app.Use(uweb.MdLogger(uweb.LOG_LEVEL_2))
	
	// Cache use memcache
	app.Use(uweb.MdCache("memcache", "127.0.0.1:11211"))
	
	// Session depends on cache
	app.Use(uweb.MdSession(3600*12))
	
	// Flash depends on session
	app.Use(uweb.MdFlash())
	
	// Csrf depends on session, and get the Csrf token from session with key: CSRF_TOKEN_KEY
	app.Use(uweb.MdCsrf())
	
	// Read ini config file
	app.Use(uweb.MdConfig("../../etc/demo.cfg"))
	
	// Html render
	app.Use(uweb.MdRender("../../pub/html"))
	
	// Cors
	app.Use(uweb.MdCors(uweb.DefaultCors))
	
	// if you want more method, change route.go
	app.Use(uweb.MdRouter())
	
	// listen address
	app.Listen(":9099")
}

//
// src/ctrls/auth/login.go 
//
package auth

import (
	   "github.com/ot24net/uweb"
	   "models/account"
)

func init() {
	 // simple get
	 uweb.Get("/account/login", func(c *uweb.Context) {
	 	 data := map[string]string {
	 	 	  "key": "value"
		 }		  	  
	 	 c.Render.Html("account/login.html", data)
	 })	
	 
	 // not support regexp match
	 uweb.Put("/account/:user_id", func (c *uweb.Context) {
	     userId := c.Req.Params["user_id"]
	 	 println(userId)
	 	 account.Noop(userId)
	 	 c.Res.Status = 201
	 	 c.Render.Plain("success")
     })
}

```

## Design
There is middleware system, but if want to extend, change the source code.

## Performance
Route middleware is rather fast, especially for long path, as it stores paths in tree. 
Session middleware depends on cache, which will slow down the benchmark.
