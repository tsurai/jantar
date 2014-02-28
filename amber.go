// Package amber is a minimalist mvc web framework written in golang.
//
// It has been largely inspired by [Martini](https://github.com/codegangsta/martini) but emphasizes performance over
// syntactic sugar by avoiding unnecessary reflections.
//
// While trying to be slim amber provides a number of security features out of the box to avoid common security vulnerabilities.
package amber

import (
  "amber/context"
  "os"
  "log"
  "fmt"
  "time"
  "net/http"
)

// logger is a package global logger instance using the prefix "[amber] " on outputs
var (
  logger *log.Logger
)

// Amber is the top level application type
type Amber struct {
  hostname    string
  port        uint
  middleware  []IMiddleware
  tm          *TemplateManager
  *router
}

// New creates a new Amber instance ready to listen on a given hostname and port
func New(hostname string, port uint) *Amber {
  a := &Amber{
    hostname: hostname,
    port: port,
    tm: newTemplateManager("views"),
    router: newRouter(hostname, port),
    middleware: nil,
  }

  logger = log.New(os.Stdout, "[amber] ", 0)
  context.SetGlobal("TemplateManager", a.tm)
  context.SetGlobal("Router", a.router)

  a.AddRoute("GET", "/public/.+", servePublic)

  return a
}

// AddMiddleware adds a given middleware to the current middleware list. Middlewares are executed
// once for every request before the actual route handler is called
func (a *Amber) AddMiddleware(mware IMiddleware) {
  if len(a.middleware) > 0 {
    a.middleware[len(a.middleware)-1].setNext(&mware)
  }
  a.middleware = append(a.middleware, mware)
}

func (a *Amber) initMiddleware() {
  for _, mw := range a.middleware {
    mw.Initialize()
  }
}

func (a *Amber) callMiddleware(respw http.ResponseWriter, req *http.Request) bool {
  for _, mw := range a.middleware {
    if !mw.Call(respw, req) {
      return false
    }
    
    if mw.doesYield() {
      break
    }
  }
  return true  
}

// ServeHTTP implements the http.Handler interface
func (a *Amber) ServeHTTP(respw http.ResponseWriter, req *http.Request) {
  t0 := time.Now()

  if method := req.FormValue("_method"); method != "" {
    req.Method = method
  }

  logger.Printf("%s %s", req.Method, req.RequestURI)

  if route := a.searchRoute(req); route != nil {
    if a.callMiddleware(respw, req) {
      route.handler(respw, req)
    }
  } else {
    logger.Printf("404 page not found")
    http.NotFound(respw, req)
  }
  logger.Printf("Completed in %v", time.Since(t0))
}

// Run starts the http server and listens on the hostname and port given to New.
func (a *Amber) Run() {
  a.initMiddleware()

  if err := a.tm.loadTemplates(); err != nil {
    logger.Fatal("[Fatal]", err)
  }
  
  logger.Println("Starting server & listening on port", a.port)
  logger.Fatal("[Fatal]", http.ListenAndServe(fmt.Sprintf(":%d", a.port), a))
}
