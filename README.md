#uweb Web Framework

uweb is a web framework written in Golang. 
It borrows many ideas from Koa.js, Gin, Playframework, etc.

## example
```
//
// src/webapp/app/main.go
// 
package main

import (
	"github.com/ot24net/uweb"
	
	_ "webapp/ctrl"
)

func main() {
	// uweb
	uweb.DEBUG = true
	uweb.DEVELOPMENT = true // will reload html template on each access
	uweb.SID_COOKIE_KEY = "_uweb_sid"

	// app
	app := uweb.NewApp()

	// hacheck
	app.Use(uweb.MdIgnore([]string{"/hacheck"}))

	// static
	app.Use(uweb.MdFavicon("../public/img/favicon.ico"))
	app.Use(uweb.MdStatic("/public", "../public")) // before compress

	// compress
	app.Use(uweb.MdCompress())

	// log
	app.Use(uweb.MdLogger(uweb.LOG_LEVEL_2))

	// session
	app.Use(uweb.MdCache("memcache", "localhost:11211"))
	app.Use(uweb.MdSession(3600 * 24 * 14))
	app.Use(uweb.MdFlash())

	// csrf
	app.Use(uweb.MdCsrf())

	// render
	app.Use(uweb.MdRender("../public/html", ".html"))

	// redirect
	app.Use(uweb.MdRedirect())

	// error page
	app.Use(uweb.MdErrPage(uweb.Map{
		"404_home_url": "http://goto_myhost.com",
	}))

	// router
	app.Use(uweb.MdRouter())

	// listen
	app.Listen(":9090")
}

//
// src/webapp/ctrl/index.go
//
package ctrl

import (
	   _ "webapp/ctrl/demo1"
	   // import other module
)

//
// src/webapp/ctrl/demo1/demo1.go
//
package demo1

import (
	   "github.com/ot24net/uweb"
	   "webapp/model/demo1"
)

func init() {
	 // simple get
	 uweb.Get("/demo1/login", func(c *uweb.Context) {
	     demo1.Noop(123)
	 	 c.Render.Html(200, "demo1/login", uweb.Map{
		     "title": "hello demo1",
		 })
	 })	
}

//
// web/src/model/demo1/noop.go
//
package demo1

func Noop(userId int) {
	// do nothing
}

//
// web/src/public/html/demo1/login.html
//
{{define "demo1/login"}}

<!doctype html>
<html>
  <head>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
	<title>{{.title}}</title>
	<meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0">
  </head>
  <body>
      uweb test
  </body>
</html>

{{end}}

```

## Design
There is middleware system, but if want to extend, change the source code.

## Performance
Route middleware is rather fast, especially for long path, as it stores paths in tree. 
Session middleware depends on cache, which will slow down the benchmark.

## Who is using it
newding.com use it in several WeChat based web apps;
ot24.net use it in its internal admin platform;