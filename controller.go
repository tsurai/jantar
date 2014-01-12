package amber

import (
  "fmt"
  "strings"
  "reflect"
  "net/http"
  "net/url"
  "html/template"
)

type IController interface {
  Redirect(name string)
  Render()
}

type Controller struct {
  *context
  *Validation
  Session     map[string]string
  flash       map[string]string
  alert       map[string]template.HTML
  RenderArgs  map[string]interface{}
}

func newController(ctx *context) interface{} {
  con := reflect.New(ctx.route.controllerType)
  con.Elem().Field(0).Set(reflect.ValueOf(new(Controller)))

  base := con.Elem().Field(0).Interface().(*Controller)
  base.context = ctx
  base.Validation = &Validation{}
  base.Validation.errors = make(map[string][]string)
  base.RenderArgs = make(map[string]interface{})
  base.Session = make(map[string]string)
  base.flash = make(map[string]string)
  base.alert = make(map[string]template.HTML)

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
        base.alert[key] = template.HTML(val[0])
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
    logger.Println("![Warning]! Failed to parse post data. Data is nil")
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

func (c *Controller) Redirect(name string) {
  url := c.route.router.getReverseUrl(name, nil)
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
    logger.Println("![Warning]! Can't find template ", tmplName)
  } else if err := tmpl.Execute(c.rw, c.RenderArgs); err != nil {
    logger.Println("![Warning]! Failed to render template:", err.Error())
  }
}