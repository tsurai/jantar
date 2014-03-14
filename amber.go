// Jantar is a lightweight mvc web framework with emphasis on security written in golang.
//
// It has been largely inspired by Martini(https://github.com/codegangsta/martini) but prefers performance over 
// syntactic sugar and aims to provide crucial security settings and features right out of the box.
package jantar

import (
  "github.com/tsurai/jantar/context"
  "os"
  "os/signal"
  "log"
  "fmt"
  "net"
  "time"
  "sync"
  "strings"
  "net/http"
  "crypto/tls"
)

// logger is a package global logger instance using the prefix "[jantar] " on outputs
var (
  logger *log.Logger
)

const (
  TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 uint16 = 0xc02c
  TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 uint16 = 0xc030
)

// Jantar is the top level application type
type Jantar struct {
  closing     bool
  wg          sync.WaitGroup
  listener    net.Listener
  config      *Config
  middleware  []IMiddleware
  tm          *TemplateManager
  router      *router
}

type TlsConfig struct {
  CertFile    string
  KeyFile     string
  CertPem     []byte
  KeyPem      []byte
  cert        tls.Certificate
}

type Config struct {
  Hostname    string
  Port        int
  Tls         *TlsConfig
}

// New creates a new Jantar instance ready to listen on a given hostname and port
func New(config *Config) *Jantar {
  if config == nil {
    logger.Fatal("[Fatal] No config given")
  }

  a := &Jantar{
    config: config,
    tm: newTemplateManager("views"),
    router: newRouter(),
    middleware: nil,
    closing: false,
  }

  if a.config.Port < 1 {
    if a.config.Tls == nil {
      a.config.Port = 80
    } else {
      a.config.Port = 443
    }
  }

  // create logger
  logger = log.New(os.Stdout, "[jantar] ", 0)

  // load default middleware
  a.AddMiddleware(&csrf{})

  // load ssl certificate
  if config.Tls != nil {
    a.loadCertificate()
  }
  
  context.SetGlobal("TemplateManager", a.tm)
  context.SetGlobal("Router", a.router)

  a.AddRoute("GET", "/public/.+", servePublic)

  return a
}

func (a *Jantar) loadCertificate() {
  var err error
  var cert tls.Certificate
  conf := a.config.Tls

  if conf.CertFile != "" && conf.KeyFile != "" {
    cert, err = tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
  } else if conf.CertPem != nil && conf.KeyPem != nil {
    cert, err = tls.X509KeyPair(conf.CertPem, conf.KeyPem)
  } else {
    logger.Fatal("[Fatal] Can't load X509 certificate. Reason: Missing parameter")
  }

  if err != nil {
    logger.Fatal("[Fatal] Can't load X509 certificate. Reason: ", err)
  }

  a.config.Tls.cert = cert
}

// AddMiddleware adds a given middleware to the current middleware list. Middlewares are executed
// once for every request before the actual route handler is called
func (a *Jantar) AddMiddleware(mware IMiddleware) {
  if len(a.middleware) > 0 {
    a.middleware[len(a.middleware)-1].setNext(&mware)
  }
  a.middleware = append(a.middleware, mware)
}

func (a *Jantar) initMiddleware() {
  for _, mw := range a.middleware {
    mw.Initialize()
  }
}

func (a *Jantar) cleanupMiddleware() {
 for _, mw := range a.middleware {
    mw.Cleanup()
  } 
}

func (a *Jantar) callMiddleware(respw http.ResponseWriter, req *http.Request) bool {
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

// AddRoute adds a route with given method, pattern and handler to the Router
func (a *Jantar) AddRoute(method string, pattern string, handler interface{}) *route {
  return a.router.addRoute(method, pattern, handler)
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

func (a *Jantar) listenForSignals() {
  sigChan := make(chan os.Signal, 1)

  signal.Notify(sigChan, os.Interrupt, os.Kill)

  s := <-sigChan
  if s == os.Kill {
    logger.Println("[Fatal] Got SIGKILL")
  }

  a.Stop()
}

func (a *Jantar) listenAndServe(addr string, handler http.Handler) error {
  if addr == "" {
    addr = ":http"
  }

  server := &http.Server{Addr: addr, Handler: handler}
  
  var err error

  if a.config.Tls != nil {
    // configure tls with secure settings
    a.listener, err = tls.Listen("tcp", addr, &tls.Config{
      Certificates: []tls.Certificate{a.config.Tls.cert},
      CipherSuites: []uint16{
        TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        tls.TLS_RSA_WITH_AES_256_CBC_SHA,
        tls.TLS_RSA_WITH_AES_128_CBC_SHA,
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
        tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
        tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
        tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
        tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
        tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
      },
      PreferServerCipherSuites: true,
      MinVersion: tls.VersionTLS10,
    })

    // listen redirect port 80 to 443 if using the standard port
    if a.config.Port == 443 {
      go http.ListenAndServe(fmt.Sprintf("%s:%d", a.config.Hostname, 80), http.HandlerFunc(
        func(respw http.ResponseWriter, req *http.Request) {
          http.Redirect(respw, req, "https://" + a.config.Hostname + req.RequestURI, 301)
        }))
    }
   } else {
   a.listener, err = net.Listen("tcp", addr) 
  }

  if err != nil {
    return err
  }

  if err = server.Serve(a.listener); !a.closing {
    return err
  } 

  return nil
}

// ServeHTTP implements the http.Handler interface
func (a *Jantar) ServeHTTP(respw http.ResponseWriter, req *http.Request) {
  a.wg.Add(1)

  t0 := time.Now()

  if method := req.FormValue("_method"); method != "" {
    req.Method = method
  }

  logger.Printf("%s %s", req.Method, req.RequestURI)

  // set security header
  respw.Header().Set("Strict-Transport-Security", "max-age=31536000;includeSubDomains")
  respw.Header().Set("X-Frame-Options", "sameorigin")
  respw.Header().Set("X-XSS-Protection", "1;mode=block")
  respw.Header().Set("X-Content-Type-Options", "nosniff")

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
func (a *Jantar) Stop() {
  a.closing = true

  // stop listening for new connections
  a.listener.Close()

  // wait until all pending requests have been finished
  a.wg.Wait()

  a.cleanupMiddleware()
}

// Run starts the http server and listens on the hostname and port given to New
func (a *Jantar) Run() {
  a.initMiddleware()

  if err := a.tm.loadTemplates(); err != nil {
    logger.Fatal("[Fatal]", err)
  }

  go a.listenForSignals()
  
  logger.Println("Starting server & listening on port", a.config.Port)
  
  if err := a.listenAndServe(fmt.Sprintf("%s:%d", a.config.Hostname, a.config.Port), a); err != nil {
    logger.Println(err)
  }
  
  logger.Println("Stopping server")
}