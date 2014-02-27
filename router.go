package amber


import (
  "amber/context"
  "os"
  "fmt"
  "strconv"
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
  router          *router
}

type router struct {
  hostname    string
  port        string
  namedRoutes map[string]*route
  routes      []*route
}

// Router functions ----------------------------------------------
func newRouter(hostname string, port uint) *router {
  return &router{hostname: hostname, port: strconv.FormatUint(uint64(port), 10), namedRoutes: make(map[string]*route)}
}

func (r *router) AddRoute(method string, pattern string, handler interface{}) *route {
  route := newRoute(strings.ToUpper(method), pattern, handler, r)
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

    if r.port != "80" || r.port != "8080" {
      return "http://" + r.hostname + ":" + r.port + url
    }
    return "http://" + r.hostname + url
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
func newRoute(method string, pattern string, handler interface{}, router *router) *route {
  var finalFunc http.HandlerFunc
  cName := ""
  cAction := ""

  if reflect.TypeOf(handler) == reflect.TypeOf(http.NotFound) {
    finalFunc = handler.(func(http.ResponseWriter, *http.Request))
  } else if cType := getControllerType(handler); cType != nil {
    fn := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
    if fn == nil {
      logger.Println("[Warning] Failed to add route. Can't fetch controller function")
      return nil
    }

    token := strings.Split(fn.Name(), ".")
    cName = token[1][2:len(token[1])-1]
    cAction = token[2]

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

  return &route{cName, cAction, pattern, method, finalFunc, regexp.MustCompile(regexPattern), router}
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