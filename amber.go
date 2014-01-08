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
	tm				*templateManager
	*router
}

func New() *Amber {
	router := newRouter()

	a := &Amber{
		port: 3000,
		tm: newtemplateManager("views", router),
		router: router,
	}

	logger = log.New(os.Stdout, "[amber] ", 0)
	a.tm.loadTemplates()

	a.AddRouteFunc("GET", "/public/.+", servePublic)

	return a
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
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", a.port), a))
}