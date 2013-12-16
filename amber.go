package amber

import (
	"os"
	"log"
	"fmt"
	"time"
	"net/http"
)

type Handler interface{}
type Param map[string]string

var logger *log.Logger

type Amber struct {
	port 			uint
	handlers 	[]http.Handler
	tm				*tmplManager
	IRouter
}

func New() *Amber {
	a := &Amber{
		port: 3000,
		tm: &tmplManager{
			directory: "views",
		},
		IRouter: NewRouter(),
	}

	logger = log.New(os.Stdout, "[amber] ", 0)
	a.tm.loadTemplates()

	a.AddRoute("GET", "/public/.+", servePublic)

	return a
}

func (a *Amber) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
  t0 := time.Now()
	logger.Printf("%s %s", req.Method, req.RequestURI)

	if r, p := a.SearchRoute(req.Method, req.URL.Path); r != nil {
		newContext(r.handler, r, rw, req, p).callHandler()
	} else {
		http.NotFound(rw, req)
	}

	logger.Printf("Completed in %v", time.Since(t0))	
}

func (a *Amber) Run() {
	logger.Println("Starting server & listening on port", a.port)
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", a.port), a))
}