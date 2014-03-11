// Package amber is a minimalist mvc web framework written in golang.
//
// It has been largely inspired by Martini(https://github.com/codegangsta/martini) but emphasizes performance over
// syntactic sugar by avoiding unnecessary reflections.
//
// While trying to be slim amber provides a number of security features out of the box to avoid common security vulnerabilities.
package amber

import (
  "github.com/tsurai/amber/context"
  "os"
  "os/signal"
  "log"
  "fmt"
  "net"
  "time"
  "sync"
  "strings"
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
  closing     bool
  wg          sync.WaitGroup
  listener    net.Listener
  middleware  []IMiddleware
  tm          *TemplateManager
  router      *router
}

// New creates a new Amber instance ready to listen on a given hostname and port
func New(hostname string, port uint) *Amber {
  a := &Amber{
    hostname: hostname,
    port: port,
    tm: newTemplateManager("views"),
    router: newRouter(hostname, port),
    middleware: nil,
    closing: false,
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

func (a *Amber) AddRoute(method string, pattern string, handler interface{}) *route {
  return a.router.addRoute(method, pattern, handler)
}


func (a *Amber) setSecurityHeader(header http.Header) {
  header.Add("X-Content-Type-Options", "nosniff")
}

func servePublic(rw http.ResponseWriter, req *http.Request) {
  var file http.File
  var err error
  var stat os.FileInfo
  fname := req.URL.Path[len("/public/"):]

  if !strings.HasPrefix(fname, ".") {
    if file, err = http.Dir("public").Open(fname); err == nil {
      if stat, err = file.Stat(); err == nil {
        if !stat.IsDir() {
          http.ServeContent(rw, req, req.URL.Path, stat.ModTime(), file)
          file.Close()
          return
        }
      }
    }
  }

  http.NotFound(rw, req)
}

func (a *Amber) listenForSignals() {
  sigChan := make(chan os.Signal, 1)

  signal.Notify(sigChan, os.Interrupt, os.Kill)

  s := <-sigChan
  if s == os.Kill {
    logger.Println("[Fatal] Got SIGKILL")
  }

  a.Stop()
}

func (a *Amber) listenAndServe(addr string, handler http.Handler) error {
  if addr == "" {
    addr = ":http"
  }

  server := &http.Server{Addr: addr, Handler: handler}
  
  var err error
  a.listener, err = net.Listen("tcp", addr)
  if err != nil {
    return err
  }

  if err = server.Serve(a.listener); !a.closing {
    return err
  } 

  return nil
}

// ServeHTTP implements the http.Handler interface
func (a *Amber) ServeHTTP(respw http.ResponseWriter, req *http.Request) {
  a.wg.Add(1)

  t0 := time.Now()

  if method := req.FormValue("_method"); method != "" {
    req.Method = method
  }

  logger.Printf("%s %s", req.Method, req.RequestURI)

  if route := a.router.searchRoute(req); route != nil {
    if a.callMiddleware(respw, req) {
      route.handler(respw, req)
    }
  } else {
    logger.Printf("404 page not found")
    http.NotFound(respw, req)
  }

  context.ClearData(req)
  logger.Printf("Completed in %v", time.Since(t0))

  a.wg.Done()
}

// Stop closes the listener and stops the server when all pending requests have been finished
func (a *Amber) Stop() {
  a.closing = true

  // stop listening for new connections
  a.listener.Close()

  // wait until all pending requests have been finished
  a.wg.Wait()
}

// Run starts the http server and listens on the hostname and port given to New
func (a *Amber) Run() {
  a.initMiddleware()

  if err := a.tm.loadTemplates(); err != nil {
    logger.Fatal("[Fatal]", err)
  }

  go a.listenForSignals()
  
  logger.Println("Starting server & listening on port", a.port)
  
  if err := a.listenAndServe(fmt.Sprintf(":%d", a.port), a); err != nil {
    logger.Println(err)
  }
  
  logger.Println("Stopping server")
}