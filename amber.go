package amber

import (
  "os"
  "log"
  "fmt"
  "time"
  "reflect"
  "net/http"
)

type Handler interface{}
type Param map[string]string

var logger *log.Logger

type Amber struct {
  hostname  string
  port      uint
  handlers  []http.Handler
  tm        *templateManager
  *router
}

func New(hostname string, port uint) *Amber {
  router := newRouter(hostname, port)

  a := &Amber{
    hostname: hostname,
    port: port,
    tm: newTemplateManager("views", router),
    router: router,
  }

  logger = log.New(os.Stdout, "[amber] ", 0)
  a.tm.loadTemplates()

  a.AddRouteFunc("GET", "/public/.+", servePublic)

  return a
}

func (a *Amber) AddModule(config interface{}) {
  var err error

  switch reflect.TypeOf(config) {
  case reflect.TypeOf(&MailerConfig{}):
    err = Mailer.initialize(a.tm, config.(*MailerConfig))
  case reflect.TypeOf(&DatabaseConfig{}):
    err = DB.initialize(config.(*DatabaseConfig))
  }

  if err != nil {
    logger.Fatal("[Fatal]", err)
  }
}

func (a *Amber) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
  t0 := time.Now()
  logger.Printf("%s %s", req.Method, req.RequestURI)

  if route, param := a.searchRoute(req.Method, req.RequestURI); route != nil {
    if route.isController {
      newContext(route.handler, route, rw, req, a.tm, param).callHandler()
    } else {
      route.handler.(func(http.ResponseWriter, *http.Request))(rw, req)
    }
  } else {
    http.NotFound(rw, req)
  }

  logger.Printf("Completed in %v", time.Since(t0))  
}

func (a *Amber) Run() {
  logger.Println("Starting server & listening on port", a.port)
  logger.Fatal("[Fatal]", http.ListenAndServe(fmt.Sprintf(":%d", a.port), a))
}