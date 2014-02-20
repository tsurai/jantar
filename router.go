package amber


import (
  "os"
  "fmt"
  "strconv"
  "strings"
  "regexp"
  "net/http"
  "amber/context"
)

type route struct {
  pattern         string
  method          string
  handler         http.HandlerFunc
  router          *router
  regex           *regexp.Regexp
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

func (r *router) AddRoute(method string, pattern string, handler http.HandlerFunc) *route {
  route := newRoute(method, pattern, handler, r)
  r.routes = append(r.routes, route)

  return route
}

func (r *router) searchRoute(req *http.Request) *route {
  method := req.Method
  request := req.RequestURI

  for i, route := range r.routes {
    if route.method == method || method == "Any" {
      matches := route.regex.FindStringSubmatch(request)
      if len(matches) > 0 && matches[0] == request {
        params := make(map[string]string)
        for i := 1; i < len(matches)-1; i++ {
          params[route.regex.SubexpNames()[i]] = matches[i]
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
func newRoute(method string, pattern string, handler http.HandlerFunc, router *router) *route {
  regex := regexp.MustCompile("{[a-zA-Z0-9]+}")
  regexPattern := regex.ReplaceAllStringFunc(pattern, func(s string) string {
    return fmt.Sprintf("(?P<%s>[^/]+)", s[1:len(s)-1])
  })
  regexPattern = regexPattern + "/?(\\?.*)?"

  return &route{pattern, method, handler, router, regexp.MustCompile(regexPattern)}
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