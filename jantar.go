// Package jantar is a lightweight mvc web framework with emphasis on security written in golang.
//
// It has been largely inspired by Martini(https://github.com/codegangsta/martini) but prefers performance over
// syntactic sugar and aims to provide crucial security settings and features right out of the box.
package jantar

import (
	"crypto/tls"
	"fmt"
	"github.com/tsurai/jantar/context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Log is a package global Log instance using the prefix "[jantar] " on outputs
var (
	Log *JLogger
)

// Jantar is the top level application type
type Jantar struct {
	closing    bool
	wg         sync.WaitGroup
	listener   net.Listener
	config     *Config
	middleware []IMiddleware
	tm         *TemplateManager
	router     *router
}

// TLSConfig can be given to Jantar to enable tls support
type TLSConfig struct {
	CertFile string
	KeyFile  string
	CertPem  []byte
	KeyPem   []byte
	cert     tls.Certificate
}

// Config is the main configuration struct for jantar
type Config struct {
	Hostname string
	Port     int
	TLS      *TLSConfig
}

// New creates a new Jantar instance ready to listen on a given hostname and port.
// Choosing a port small than 1 will cause Jantar to use the standard ports.
func New(config *Config) *Jantar {
	// create Log
	Log = NewJLogger(os.Stdout, "", LogLevelInfo)

	if config == nil {
		Log.Fatal("no config given")
	}

	j := &Jantar{
		config:     config,
		tm:         newTemplateManager("views"),
		router:     newRouter(),
		middleware: nil,
		closing:    false,
	}

	if j.config.Port < 1 {
		if j.config.TLS == nil {
			j.config.Port = 80
		} else {
			j.config.Port = 443
		}
	}

	// load default middleware
	j.AddMiddleware(&csrf{})

	// load ssl certificate
	if config.TLS != nil {
		if err := loadTLSCertificate(config.TLS); err != nil {
			Log.Fatald(JLData{"error": err}, "failed to load x509 certificate")
		}
	}

	setModule(ModuleTemplateManager, j.tm)
	setModule(ModuleRouter, j.router)

	return j
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
	var err error

	if addr == "" {
		addr = ":http"
	}

	if j.config.TLS != nil {
		// configure tls with secure settings
		tlsConfig.Certificates = []tls.Certificate{j.config.TLS.cert}
		j.listener, err = tls.Listen("tcp", addr, tlsConfig)

		// listen redirect port 80 to 443 if using the standard port
		if j.config.Port == 443 {
			go http.ListenAndServe(fmt.Sprintf("%s:%d", j.config.Hostname, 80), http.HandlerFunc(
				func(respw http.ResponseWriter, req *http.Request) {
					http.Redirect(respw, req, "https://"+j.config.Hostname+req.RequestURI, 301)
				}))
		}
	} else {
		j.listener, err = net.Listen("tcp", addr)
	}

	if err != nil {
		return err
	}

	server := &http.Server{Addr: addr, Handler: handler}
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

	Log.Infof("%s %s", req.Method, req.URL.Path)

	// set security header
	respw.Header().Set("Strict-Transport-Security", "max-age=31536000;includeSubDomains")
	respw.Header().Set("X-Frame-Options", "sameorigin")
	respw.Header().Set("X-XSS-Protection", "1;mode=block")
	respw.Header().Set("X-Content-Type-Options", "nosniff")

	context.Set(req, "_RenderArgs", make(map[string]interface{}), true)
	if j.callMiddleware(respw, req) {
		if route := j.router.searchRoute(req); route != nil {
			route.handler(respw, req)
		} else {
			ErrorHandler(http.StatusNotFound)(respw, req)
		}
	}

	context.ClearData(req)
	Log.Infof("completed in %v", time.Since(t0))

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
		Log.Fatal("failed to load templates")
	}

	go j.listenForSignals()

	Log.Infod(JLData{"hostname": j.config.Hostname, "port": j.config.Port, "TLS": j.config.TLS != nil}, "starting server & listening")

	if err := j.listenAndServe(fmt.Sprintf("%s:%d", j.config.Hostname, j.config.Port), j); err != nil {
		Log.Fatal(err)
	}

	Log.Info("stopping server")
}
