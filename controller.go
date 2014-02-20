package amber

import (
  "strings"
  "runtime"
  "reflect"
  "net/http"
  "amber/context"
)

type IController interface {
  setInternal(rw http.ResponseWriter, r *http.Request, name string, action string)
  Render()
}

type Controller struct {
  Respw   http.ResponseWriter
  Req     *http.Request
  name    string
  action  string
}

func newController(t reflect.Type, respw http.ResponseWriter, req *http.Request, name string, action string) IController {
  c := reflect.New(t).Interface().(IController)
  c.setInternal(respw, req, name, action)

  return c
}

func getControllerType(handler interface{}) reflect.Type {
  t := reflect.TypeOf(handler)
  if t.Kind() == reflect.Func && t.NumIn() != 0 && t.In(0).Implements(reflect.TypeOf((*IController)(nil)).Elem()) {
    return t.In(0).Elem()
  }

  logger.Fatalf("[Fatal] Handler is no valid controller function '%v'", t)
  return nil
}


func CallController(ctrlFunc interface{}) (func(rw http.ResponseWriter, r *http.Request)) {
  t := getControllerType(ctrlFunc)

  fn := runtime.FuncForPC(reflect.ValueOf(ctrlFunc).Pointer());
  if fn == nil {
    logger.Fatal("[Fatal] Failed to add route. Can't fetch controller function")
  }

  token := strings.Split(fn.Name(), ".")
  cName := token[1][2:len(token[1])-1]
  cAction := token[2]
  
  return func(rw http.ResponseWriter, r *http.Request) {
    c := newController(t, rw, r, cName, cAction)

    var in []reflect.Value
    in = append(in, reflect.ValueOf(c))

    reflect.ValueOf(ctrlFunc).Call(in)
  }
}

func (c *Controller) setInternal(respw http.ResponseWriter, req *http.Request, name string, action string) {
  c.Respw = respw
  c.Req = req
  c.name = name
  c.action = action
}

func (c *Controller) Render() {
  tm := context.GetGlobal("TemplateManager").(*TemplateManager)
  
  tmplName := c.name + "/" + c.action + ".html"

  args := make(map[string]interface{})
  args["csrf"] = context.Get(c.Req, "csrf");
  
  if err := tm.RenderTemplate(c.Respw, tmplName, args); err != nil {
    logger.Println(err.Error())
  }
}
/*
func CallHandler(rw http.ResponseWriter, r *http.Request) {
  var in []reflect.Value

  c := newController(rw, r)
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

/*
type IController interface {
  Render()
}

type Controller struct {
  *Context
  *Validation
  Session     map[string]string
  flash       map[string]string
  alert       map[string]string
}

func newController(ctx *Context) interface{} {
  con := reflect.New(ctx.route.controllerType)
  con.Elem().Field(0).Set(reflect.ValueOf(new(Controller)))

  base := con.Elem().Field(0).Interface().(*Controller)
  base.Context = ctx
  base.Validation = &Validation{}
  base.Validation.errors = make(map[string][]string)
  base.Session = make(map[string]string)
  base.flash = make(map[string]string)
  base.alert = make(map[string]string)

  // fetch validation errors from cookie

  if cookie, err := ctx.Req.Cookie("AMBER_ERRORS"); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      base.Validation.errors = m
    }

    // delete cookie
    cookie.MaxAge = -9999
    http.SetCookie(ctx.rw, cookie)
  }

  // fetch flash from cookie
  if cookie, err := ctx.Req.Cookie("AMBER_FLASH"); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      for key, val := range m {
        base.flash[key] = val[0]
      }
    }

    // delete cookie
    cookie.MaxAge = -9999
    http.SetCookie(ctx.rw, cookie)
  }

  // fetch alert from cookie
  if cookie, err := ctx.Req.Cookie("AMBER_ALERT"); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      for key, val := range m {
        base.alert[key] = val[0]
      }
    }

    // delete cookie
    cookie.MaxAge = -9999
    http.SetCookie(ctx.rw, cookie)
  }

  // fetch session from cookie
  if cookie, err := ctx.Req.Cookie("AMBER_SESSION"); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      for key, val := range m {
        base.Session[key] = val[0]
      }
    }
  }

  return con.Interface()
}

func isControllerHandler(handler Handler) (bool, reflect.Type) {
  t := reflect.TypeOf(handler)
  if t.Kind() == reflect.Func && t.NumIn() != 0 && t.In(0).Implements(reflect.TypeOf((*IController)(nil)).Elem()) {
    return true, t.In(0).Elem()
  }
  return false, nil
}


// WARNING: this is case sensitive! Using the wrong case in the html form can cause misbehavior
func (c *Controller) ExtractObject(name string, obj interface{}) interface{} {
  if len(c.PostParam) <= 0 {
    logger.Println("[Warning] Failed to parse post data. Data is nil")
    return nil
  }

  if reflect.TypeOf(obj).Kind() != reflect.Ptr {
    return nil
  }

  objvalue := reflect.ValueOf(obj)

  for key, value := range c.PostParam {
    substr := strings.SplitN(key, ".", 2)
    if (len(substr) == 2) && (strings.EqualFold(substr[0], name)) {
      objvalue.Elem().FieldByName(substr[1]).Set(reflect.ValueOf(value))
      delete(c.PostParam, key)
    }
  }

  return objvalue.Interface()
}

func (c *Controller) AddFlash(name string, obj interface{}) {
  t := reflect.TypeOf(obj)

  if t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
    value := reflect.ValueOf(obj)
    for i := 0; i < t.NumField(); i++ {
      field := t.Field(i)
      if field.Tag.Get("amber") != "noflash" {
        c.flash[name + "." + field.Name] = fmt.Sprintf("%v", value.Field(i).Interface())
      }
    }
  } else {
    // TODO: there has to be a better way
    c.flash[name] = fmt.Sprintf("%v", obj)
  }
}

func (c *Controller) AlertSuccess(msg string) {
  http.SetCookie(c.rw, &http.Cookie{Name: "AMBER_ALERT", Value: url.Values{"success": []string{msg}}.Encode(), Secure: false, HttpOnly: true, Path: "/"})
}

func (c *Controller) AlertError(msg string) {
  http.SetCookie(c.rw, &http.Cookie{Name: "AMBER_ALERT", Value: url.Values{"danger": []string{msg}}.Encode(), Secure: false, HttpOnly: true, Path: "/"})
}

func (c *Controller) SaveFlash() {
  if len(c.flash) > 0 {
    values := url.Values{}
    for key, val := range c.flash {
      values.Add(key, val)
    }

    http.SetCookie(c.rw, &http.Cookie{Name: "AMBER_FLASH", Value: values.Encode(), Secure: false, HttpOnly: true, Path: "/"})
  }
}

func (c *Controller) SaveErrors() {
  if c.Validation.hasErrors {
    values := url.Values{}
    for key, array := range c.Validation.errors {
      for _, val := range array {
        values.Add(key, val)
      }
    }

    http.SetCookie(c.rw, &http.Cookie{Name: "AMBER_ERRORS", Value: values.Encode(), Secure: false, HttpOnly: true, Path: "/"})
  }
}

func (c *Controller) GetCookieValues(name string) url.Values {
  if cookie, err := c.Req.Cookie(name); err == nil {
    if values, err := url.ParseQuery(cookie.Value); err == nil {
      return values
    }
  }
  return nil
}

func (c *Controller) DeleteCookie(name string) {
  if cookie, err := c.Req.Cookie(name); err == nil {
    cookie.MaxAge = -9999
    http.SetCookie(c.rw, cookie)
  }
}

func (c *Controller) SaveSessionCookie(value url.Values) {
  http.SetCookie(c.rw, &http.Cookie{Name: "AMBER_SESSION", Value: value.Encode(), Secure: false, HttpOnly: true, Path: "/"}) 
}

func (c *Controller) Redirect(name string, args ...interface{}) {
  url := c.route.router.getReverseUrl(name, args)
  c.rw.Header().Set("Location", url)
  c.rw.WriteHeader(302)
}

// stupid name....
func (c *Controller) HardRedirect(url string) {
  c.rw.Header().Set("Location", url)
  c.rw.WriteHeader(302)
}

func (c *Controller) Render() {
  tmplName := c.route.cName + "/" + c.route.cAction + ".html"
  tmpl := c.tm.getTemplate(tmplName)

  if len(c.Validation.errors) > 0 {
    c.RenderArgs["errors"] = c.Validation.errors
  }

  c.RenderArgs["flash"] = c.flash
  c.RenderArgs["alert"] = c.alert

  if tmpl == nil {
    c.rw.Write([]byte("Can't find template " + tmplName))
    logger.Println("[Warning] Can't find template ", tmplName)
  } else if err := tmpl.Execute(c.rw, c.RenderArgs); err != nil {
    logger.Println("[Warning] Failed to render template:", err.Error())
  }
}*/