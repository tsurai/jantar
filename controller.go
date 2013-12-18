package amber

import (
  "reflect"
)

type IController interface {
  Render(extraArgs... Handler)
}

type Controller struct {
  *context
  Param     map[string]string
  Flash     map[string]string
}

func newController(ctx *context) Handler {
  con := reflect.New(reflect.TypeOf(ctx.handler).In(0).Elem())
  con.Elem().Field(0).Set(reflect.ValueOf(new(Controller)))

  base := con.Elem().Field(0).Interface().(*Controller)
  base.context = ctx
  base.Param = make(map[string]string)

  return con.Interface()
}

func isControllerHandler(handler Handler) bool {
  t := reflect.TypeOf(handler)
  if t.Kind() == reflect.Func && t.NumIn() != 0 && t.In(0).Implements(reflect.TypeOf((*IController)(nil)).Elem()) {
    return true
  }
  return false
}

func (c *Controller) Render(extraArgs... Handler) {
  tmplName := c.route.cName + "/" + c.route.cAction + ".html"
  tmpl := c.tm.getTemplate(tmplName)

  if tmpl == nil {
    c.rw.Write([]byte("Can't find template " + tmplName))
    logger.Println("Can't find template ", tmplName)
  } else if err := tmpl.Execute(c.rw, c.Param); err != nil {
    logger.Println("Failed to render template:", err.Error())
  }
}