package amber

import (
  "os"
  "log"
  "fmt"
  "time"
  "net/http"
  "amber/context"
)

var (
  logger *log.Logger
)

type Amber struct {
  hostname    string
  port        uint
  middleware  []IMiddleware
  tm          *TemplateManager
  *router
}

func New(hostname string, port uint) *Amber {
  router := newRouter(hostname, port)
  
  a := &Amber{
    hostname: hostname,
    port: port,
    tm: newTemplateManager("views"),
    router: router,
    middleware: nil,
  }

  logger = log.New(os.Stdout, "[amber] ", 0)
  context.SetGlobal("TemplateManager", a.tm)
  context.SetGlobal("Router", a.router)

  a.AddRoute("GET", "/public/.+", servePublic)

  return a
}

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

func (a *Amber) ServeHTTP(respw http.ResponseWriter, req *http.Request) {
  t0 := time.Now()

  logger.Printf("%s %s", req.Method, req.RequestURI)
  
  if route := a.searchRoute(req); route != nil {
    for _, mw := range a.middleware {
      mw.Call(respw, req)
      if mw.doesYield() {
        break
      }
    }
    route.handler(respw, req)
  } else {
    logger.Printf("404 Not Found")
    http.NotFound(respw, req)
  }
  logger.Printf("Completed in %v", time.Since(t0))
}

func (a *Amber) Run() {
  a.initMiddleware()

  if err := a.tm.loadTemplates(); err != nil {
    logger.Fatal("[Fatal]", err)
  }
  
  logger.Println("Starting server & listening on port", a.port)
  logger.Fatal("[Fatal]", http.ListenAndServe(fmt.Sprintf(":%d", a.port), a))
}