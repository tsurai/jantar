package amber

import (
  "strings"
  "reflect"
  "net/http"
  "net/url"
)

type IController interface {
  Redirect(name string)
  Render()
}

type Controller struct {
  *context
  *Validation
  RenderArgs  map[string]interface{}
  Flash       map[string]string
}

func newController(ctx *context) interface{} {
  con := reflect.New(ctx.route.controllerType)
  con.Elem().Field(0).Set(reflect.ValueOf(new(Controller)))

  base := con.Elem().Field(0).Interface().(*Controller)
  base.context = ctx
  base.Validation = &Validation{}
  base.Validation.errors = make(map[string][]string)
  base.RenderArgs = make(map[string]interface{})

  // fetch validation errors from cookie
  if cookie, err := ctx.Req.Cookie("AMBER_ERRORS"); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      base.Validation.errors = m
    }

    // delete cookie
    cookie.MaxAge = -9999
    http.SetCookie(ctx.rw, cookie)
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

func (c *Controller) ExtractObject(name string, obj interface{}) interface{} {
  if len(c.PostParam) <= 0 {
    logger.Println("![Warning]! Failed to parse post data. Data is nil")
    return reflect.ValueOf(nil)
  }


  if reflect.TypeOf(obj).Kind() != reflect.Ptr {
    return nil
  }

  objvalue := reflect.ValueOf(obj)

  for key, value := range c.PostParam {
    substr := strings.SplitN(key, ".", 2)
    if (len(substr) == 2) && (strings.EqualFold(substr[0], name)) {
      objvalue.Elem().FieldByName(substr[1]).Set(reflect.ValueOf(value[0]))
      delete(c.PostParam, key)
    }
  }

  return objvalue
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

  if tmpl == nil {
    c.rw.Write([]byte("Can't find template " + tmplName))
    logger.Println("![Warning]! Can't find template ", tmplName)
  } else if err := tmpl.Execute(c.rw, c.RenderArgs); err != nil {
    logger.Println("![Warning]! Failed to render template:", err.Error())
  }
}