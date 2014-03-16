package jantar


import (
  "github.com/tsurai/jantar/context"
  "fmt"
  "strings"
  "regexp"
  "reflect"
  "runtime"
  "net/http"
)

type route struct {
  cName           string
  cAction         string
  pattern         string
  method          string
  handler         http.HandlerFunc
  regex           *regexp.Regexp
}

type router struct {
  namedRoutes map[string]*route
  routes      []*route
}

// Router functions ----------------------------------------------
func newRouter() *router {
  return &router{namedRoutes: make(map[string]*route)}
}

func (r *router) addRoute(method string, pattern string, handler interface{}) *route {
  route := newRoute(strings.ToUpper(method), pattern, handler)
  r.routes = append(r.routes, route)

  // is route a controller route
  if route.cName != "" {
    // add to named routes with name as controller#action
    r.namedRoutes[strings.ToLower(route.cName+"#"+route.cAction)] = route
  }

  return route
}

func (r *router) searchRoute(req *http.Request) *route {
  method := strings.ToUpper(req.Method)
  request := req.RequestURI

  for i, route := range r.routes {
    if route.method == method || method == "ANY" {
      matches := route.regex.FindStringSubmatch(request)
      
      if len(matches) > 0 && matches[0] == request {
        params := make(map[string]string)
        
        for n := 1; n < len(matches)-1; n++ {
          params[route.regex.SubexpNames()[n]] = matches[i]
        }
        
        context.Set(req, "UrlParam", params)
        return r.routes[i]
      }
    }
  }
  return nil
}

func (r *router) getReverseUrl(name string, param []interface{}) string {
  route := r.getNamedRoute(name)
  nParam := len(param)
  
  if route != nil {
    i := -1
    regex := regexp.MustCompile("{[^/{}]+}")
    url := regex.ReplaceAllStringFunc(route.pattern, func(str string) string {
      i = i + 1
      if i < nParam {
        return fmt.Sprintf("%v", param[i])
      }
      return ""
    })

    return url
  }

  return ""
}

func (r *router) getNamedRoute(name string) *route {
  if route, ok := r.namedRoutes[strings.ToLower(name)]; ok {
    return route
  }
  
  return nil
}

// Route functions ---------------------------------------------
func newRoute(method string, pattern string, handler interface{}) *route {
  var finalFunc http.HandlerFunc
  cName := ""
  cAction := ""

  if reflect.TypeOf(handler) == reflect.TypeOf(http.NotFound) {
    finalFunc = handler.(func(http.ResponseWriter, *http.Request))
  } else if cType := getControllerType(handler); cType != nil {
    fn := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
    if fn == nil {
      logger.Warning("Failed to add route. Can't fetch controller function")
      return nil
    }

    regex := regexp.MustCompile(".*\\.\\(\\*(.*)\\)\\.(.*)")
    matches := regex.FindStringSubmatch(fn.Name())

    if len(matches) == 3 && matches[0] == fn.Name() {
      cName = matches[1]
      cAction = matches[2]
    }
    
    finalFunc = func(rw http.ResponseWriter, r *http.Request) {
      c := newController(cType, rw, r, cName, cAction)

      var in []reflect.Value
      in = append(in, reflect.ValueOf(c))

      reflect.ValueOf(handler).Call(in)
    }
  }

  regex := regexp.MustCompile("{[a-zA-Z0-9]+}")
  regexPattern := regex.ReplaceAllStringFunc(pattern, func(s string) string {
    return fmt.Sprintf("(?P<%s>[^/]+)", s[1:len(s)-1])
  })
  regexPattern = regexPattern + "/?(\\?.*)?"

  return &route{cName, cAction, pattern, method, finalFunc, regexp.MustCompile(regexPattern)}
}

func (r *route) Name(name string) {
  router := context.GetGlobal("Router").(*router)
  router.namedRoutes[strings.ToLower(name)] = r
}