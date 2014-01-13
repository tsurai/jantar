package amber

import (
  "reflect"
  "net/http"
  "net/url"
)

type context struct {
  handler Handler
  route     *route
  tm        *templateManager
  rw        http.ResponseWriter
  Req       *http.Request
  UrlParam  map[string]string
  PostParam map[string]string
  GetParam  url.Values
}

func newContext(h Handler, route *route, rw http.ResponseWriter, req *http.Request, tm *templateManager, params Param) *context {
  ctx := &context{handler: h, route: route, tm: tm, rw: rw, Req: req, UrlParam: params, PostParam: make(map[string]string)}

  var err error
  if ctx.GetParam, err = url.ParseQuery(req.URL.RawQuery); err != nil {
    logger.Println(err)
  }

  if req.Method == "POST" {
    req.ParseForm()
    for key, val := range req.Form {
      ctx.PostParam[key] = val[0]
    }
  }

  return ctx
}

func (ctx *context) callHandler() {
  var in []reflect.Value

  c := newController(ctx)
  in = append(in, reflect.ValueOf(c))

  if f, ok := reflect.TypeOf(c).MethodByName("BeforeInterceptor"); ok {
    f.Func.Call([]reflect.Value{reflect.ValueOf(c)})
  }

  // TODO: catch exception
  ret := reflect.ValueOf(ctx.handler).Call(in)
  if len(ret) > 0 {
    ctx.rw.Write([]byte(ret[0].String()))
  }
}
