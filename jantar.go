// Jantar is a lightweight mvc web framework with emphasis on security written in golang.
//
// It has been largely inspired by Martini(https://github.com/codegangsta/martini) but prefers performance over 
// syntactic sugar and aims to provide crucial security settings and features right out of the box.
package jantar

import (
  "github.com/tsurai/jantar/context"
  "os"
  "os/signal"
  "fmt"
  "net"
  "time"
  "sync"
  "strings"
  "net/http"
  "crypto/tls"
)

// Log is a package global Log instance using the prefix "[jantar] " on outputs
var (
  Log *JLogger
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

// New creates a new Jantar instance ready to listen on a given hostname and port.
// Choosing a port small than 1 will cause Jantar to use the standard ports.
func New(config *Config) *Jantar {
  // create Log
  Log = NewJLogger(os.Stdout, "", LogLevelInfo)

  if config == nil {
    Log.Fatal("No config given")
  }

  j := &Jantar{
    config: config,
    tm: newTemplateManager("views"),
    router: newRouter(),
    middleware: nil,
    closing: false,
  }

  if j.config.Port < 1 {
    if j.config.Tls == nil {
      j.config.Port = 80
    } else {
      j.config.Port = 443
    }
  }

  // load default middleware
  j.AddMiddleware(&csrf{})

  // load ssl certificate
  if config.Tls != nil {
    j.loadCertificate()
  }
  
  setModule(MODULE_TEMPLATE_MANAGER, j.tm)
  setModule(MODULE_ROUTER, j.router)

  j.AddRoute("GET", "/public/.+", servePublic)

  return j
}

func (j *Jantar) loadCertificate() {
  var err error
  var cert tls.Certificate
  conf := j.config.Tls

  if conf.CertFile != "" && conf.KeyFile != "" {
    cert, err = tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
  } else if conf.CertPem != nil && conf.KeyPem != nil {
    cert, err = tls.X509KeyPair(conf.CertPem, conf.KeyPem)
  } else {
    Log.Fatal("Failed to load X509 certificate: missing parameter")
  }

  if err != nil {
    Log.Fatald(JLData{"error": err}, "Failed to load X509 certificate", err)
  }

  j.config.Tls.cert = cert
}

// AddMiddleware adds a given middleware to the current middleware list. Middlewares are executed
// once for every request before the actual route handler is called
func (j *Jantar) AddMiddleware(mware IMiddleware) {
  if len(j.middleware) > 0 {
    j.middleware[len(j.middleware)-1].setNext(&mware)
  }
  j.middleware = append(j.middleware, mware)
}

func (j *Jantar) initMiddleware() {
  for _, mw := range j.middleware {
    mw.Initialize()
  }
}

func (j *Jantar) cleanupMiddleware() {
 for _, mw := range j.middleware {
    mw.Cleanup()
  } 
}

func (j *Jantar) callMiddleware(respw http.ResponseWriter, req *http.Request) bool {
  for _, mw := range j.middleware {
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
func (j *Jantar) AddRoute(method string, pattern string, handler interface{}) *route {
  return j.router.addRoute(method, pattern, handler)
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

func (j *Jantar) listenForSignals() {
  sigChan := make(chan os.Signal, 1)

  signal.Notify(sigChan, os.Interrupt, os.Kill)

  s := <-sigChan
  if s == os.Kill {
    Log.Fatal("Got SIGKILL")
  }

  j.Stop()
}

func (j *Jantar) listenAndServe(addr string, handler http.Handler) error {
  if addr == "" {
    addr = ":http"
  }

  server := &http.Server{Addr: addr, Handler: handler}
  
  var err error

  if j.config.Tls != nil {
    // configure tls with secure settings
    j.listener, err = tls.Listen("tcp", addr, &tls.Config{
      Certificates: []tls.Certificate{j.config.Tls.cert},
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
    if j.config.Port == 443 {
      go http.ListenAndServe(fmt.Sprintf("%s:%d", j.config.Hostname, 80), http.HandlerFunc(
        func(respw http.ResponseWriter, req *http.Request) {
          http.Redirect(respw, req, "https://" + j.config.Hostname + req.RequestURI, 301)
        }))
    }
   } else {
   j.listener, err = net.Listen("tcp", addr) 
  }

  if err != nil {
    return err
  }

  if err = server.Serve(j.listener); !j.closing {
    return err
  } 

  return nil
}

// ServeHTTP implements the http.Handler interface
func (j *Jantar) ServeHTTP(respw http.ResponseWriter, req *http.Request) {
  j.wg.Add(1)

  t0 := time.Now()

  if method := req.FormValue("_method"); method != "" {
    req.Method = method
  }

  Log.Infof("%s %s", req.Method, req.RequestURI)

  // set security header
  respw.Header().Set("Strict-Transport-Security", "max-age=31536000;includeSubDomains")
  respw.Header().Set("X-Frame-Options", "sameorigin")
  respw.Header().Set("X-XSS-Protection", "1;mode=block")
  respw.Header().Set("X-Content-Type-Options", "nosniff")

  if route := j.router.searchRoute(req); route != nil {
    if j.callMiddleware(respw, req) {
      route.handler(respw, req)
    }
  } else {
    Log.Info("404 page not found")
    http.NotFound(respw, req)
  }

  context.ClearData(req)
  Log.Infof("Completed in %v", time.Since(t0))

  j.wg.Done()
}

// Stop closes the listener and stops the server when all pending requests have been finished
func (j *Jantar) Stop() {
  j.closing = true

  // stop listening for new connections
  j.listener.Close()

  // wait until all pending requests have been finished
  j.wg.Wait()

  j.cleanupMiddleware()
}

// Run starts the http server and listens on the hostname and port given to New
func (j *Jantar) Run() {
  j.initMiddleware()

  if err := j.tm.loadTemplates(); err != nil {
    Log.Error(err)
  }

  go j.listenForSignals()

  Log.Infod(JLData{"Hostname": j.config.Hostname, "Port": j.config.Port, "TLS": j.config.Tls != nil}, "Starting server & listening")
  
  if err := j.listenAndServe(fmt.Sprintf("%s:%d", j.config.Hostname, j.config.Port), j); err != nil {
    Log.Info(err)
  }
  
  Log.Info("Stopping server")
}