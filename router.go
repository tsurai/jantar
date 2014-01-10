package amber

import (
	"os"
	"fmt"
	"strconv"
	"strings"
	"regexp"
	"net/http"
	"reflect"
	"runtime"
)

type route struct {
	pattern					string
	method					string
	isController 		bool
	controllerType 	reflect.Type
	cName						string
	cAction					string
	handler 				interface{}
	router 					*router
	regex 					*regexp.Regexp
}

type router struct {
	hostname 		string
	port 				string
	namedRoutes map[string]*route
	routes			[]*route
}

// Router functions ----------------------------------------------
func newRouter(hostname string, port uint) *router {
	return &router{hostname: hostname, port: strconv.FormatUint(uint64(port), 10), namedRoutes: make(map[string]*route)}
}

func (r *router) AddRoute(method string, pattern string, handler Handler) *route {
	route := newRoute(method, pattern, handler, r)
	if route.isController {
		route.Name(route.cName + "#" + route.cAction)
		r.routes = append(r.routes, route)

		return route
	}

	return nil
}

func (r *router) AddRouteFunc(method string, pattern string, handler func(http.ResponseWriter, *http.Request)) * route {
	route := newRoute(method, pattern, handler, r)
	r.routes = append(r.routes, route)

	return route
}

func (r *router) searchRoute(method string, request string) (*route, Param) {
	for i, route := range r.routes {
		if route.method == method || method == "Any" {
			matches := route.regex.FindStringSubmatch(request)
			if len(matches) > 0 && matches[0] == request {
				params := make(Param)
				for i := 1; i < len(matches)-1; i++ {
					params[route.regex.SubexpNames()[i]] = matches[i]
				}
				return r.routes[i], params
			}
		}
	}
	return nil, nil
}

func (r *router) getReverseUrl(name string, param []interface{}) string {
	route := r.getNamedRoute(name)
	nParam := len(param)
	
	if route != nil {
		i := -1
		regex := regexp.MustCompile("{.*}")
		url := regex.ReplaceAllStringFunc(route.pattern, func(str string) string {
			i = i + 1
			if i <= nParam - 1 {
				return param[i].(string)
			} else {
				return ""
			}
		})

		if r.port != "80" || r.port != "8080" {
			return "http://" + r.hostname + ":" + r.port + url
		} else {
			return "http://" + r.hostname + url
		}
	}

	return ""
}

func (r *router) getNamedRoute(name string) *route {
	route, ok := r.namedRoutes[strings.ToLower(name)]
	if ok {
		return route
	}
	
	return nil
}

// Route functions ---------------------------------------------
func newRoute(method string, pattern string, handler Handler, router *router) *route {
	regex := regexp.MustCompile("{[a-zA-Z0-9]+}")
	regexPattern := regex.ReplaceAllStringFunc(pattern, func(s string) string {
		return fmt.Sprintf("(?P<%s>[a-zA-Z0-9]+)", s[1:len(s)-1])
	})
	regexPattern = regexPattern + "(\\?.*)?"

	cName := ""
	cAction := ""
	isController := false
	var controllerType reflect.Type

	if ok, t := isControllerHandler(handler); ok {
		var fn *runtime.Func
	  if fn = runtime.FuncForPC(reflect.ValueOf(handler).Pointer()); fn == nil {
			logger.Println("![Warning]! Failed to add route. Can't fetch controller function")
			return nil
	  }

	  controllerType = t
	  isController = true
		token := strings.Split(fn.Name(), ".")
		cName = token[1][2:len(token[1])-1]
		cAction = token[2]
	}
	
	return &route{pattern, method, isController, controllerType, cName, cAction, handler, router, regexp.MustCompile(regexPattern)}
}

func (r *route) Name(name string) {
	r.router.namedRoutes[strings.ToLower(name)] = r
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