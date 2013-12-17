package amber

import (
	"os"
	"fmt"
	"strings"
	"regexp"
	"net/http"
	"reflect"
	"runtime"
)

type route struct {
	pattern				string
	method				string
	isController 	bool
	cName					string
	cAction				string
	router 				*router
	regex 				*regexp.Regexp
	handler 			Handler
}

type router struct {
	namedRoutes 	map[string]*route
	routes				[]*route
}

func newRouter() *router {
	return &router{namedRoutes: make(map[string]*route)}
}

func (r *router) AddRoute(method string, route string, h Handler) {
	r.routes = append(r.routes, newRoute(method, route, h, r))
}

func (r *router) searchRoute(method string, request string) (*route, Param) {
	for i, route := range r.routes {
		matches := route.regex.FindStringSubmatch(request)
		if route.method == method || method == "Any" {
			if len(matches) > 0 && matches[0] == request {
				params := make(Param)
				for i := 1; i < len(matches); i++ {
					params[route.regex.SubexpNames()[i]] = matches[i]
				}
				return r.routes[i], params
			}
		}
	}
	return nil, nil
}

func (r *route) Name(name string) {
	r.router.namedRoutes[name] = r
}

func newRoute(method string, pattern string, h Handler, r *router) *route {
	regex := regexp.MustCompile("{[a-zA-Z0-9]+}")
	pattern = regex.ReplaceAllStringFunc(pattern, func(s string) string {
		return fmt.Sprintf("(?P<%s>[a-z]+)", s[1:len(s)-1])
	})

	cName := ""
	cAction := ""
	isController := false

	if isControllerHandler(h) {
		var f *runtime.Func
	  if f = runtime.FuncForPC(reflect.ValueOf(h).Pointer()); f == nil {
			logger.Println("Failed to add route. Can't fetch controller function")
			return nil
	  }

	  isController = true
		token := strings.Split(f.Name(), ".")
		cName = token[1][2:len(token[1])-1]
		cAction = token[2]
	}
	
	return &route{pattern: pattern, method: method, handler: h, router: r, isController: isController,
								cName: cName, cAction: cAction, regex: regexp.MustCompile(pattern)}
}

func servePublic(rw http.ResponseWriter, req *http.Request) {
	var f http.File
	var err error
	var stat os.FileInfo
	fname := req.URL.Path[len("/public/"):]

	if !strings.HasPrefix(fname, ".") {
		if f, err = http.Dir("public").Open(fname); err == nil {
			if stat, err = f.Stat(); err == nil {
				if !stat.IsDir() {
					http.ServeContent(rw, req, req.URL.Path, stat.ModTime(), f)
					f.Close()
					return
				}
			}
		}
	}

	http.NotFound(rw, req)
}