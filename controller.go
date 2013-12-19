package amber

import (
  "reflect"
)

type IController interface {
  Redirect(name string)
  Render()
}

type Controller struct {
  *context
  RenderArgs  map[string]interface{}
  Flash       map[string]string
}

func newController(ctx *context) Handler {
  con := reflect.New(reflect.TypeOf(ctx.handler).In(0).Elem())
  con.Elem().Field(0).Set(reflect.ValueOf(new(Controller)))

  base := con.Elem().Field(0).Interface().(*Controller)
  base.context = ctx
  base.RenderArgs = make(map[string]interface{})

  return con.Interface()
}

func isControllerHandler(handler Handler) bool {
  t := reflect.TypeOf(handler)
  if t.Kind() == reflect.Func && t.NumIn() != 0 && t.In(0).Implements(reflect.TypeOf((*IController)(nil)).Elem()) {
    return true
  }
  return false
}

func (c *Controller) Redirect(name string) {
  url := c.route.router.getReverseUrl(name, nil)
  c.rw.Header().Set("Location", url)
  c.rw.WriteHeader(302)
}

func (c *Controller) Render() {
  tmplName := c.route.cName + "/" + c.route.cAction + ".html"
  tmpl := c.tm.getTemplate(tmplName)

  if tmpl == nil {
    c.rw.Write([]byte("Can't find template " + tmplName))
    logger.Println("![Warning]! Can't find template ", tmplName)
  } else if err := tmpl.Execute(c.rw, c.RenderArgs); err != nil {
    logger.Println("![Warning]! Failed to render template:", err.Error())
  }
}