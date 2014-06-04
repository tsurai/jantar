package jantar

import (
	"github.com/tsurai/jantar/context"
	"net/http"
	"reflect"
)

// IController describes a Controller
type IController interface {
	setInternal(rw http.ResponseWriter, r *http.Request, name string, action string)
	Render()
}

// Controller implements core functionalities of the IController interface
type Controller struct {
	name       string
	action     string
	Respw      http.ResponseWriter
	Req        *http.Request
	RenderArgs map[string]interface{}
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

	return nil
}

func (c *Controller) setInternal(respw http.ResponseWriter, req *http.Request, name string, action string) {
	c.name = name
	c.action = action
	c.Respw = respw
	c.Req = req
	c.RenderArgs = (context.Get(req, "renderArgs").(map[string]interface{}))
}

func (c *Controller) Redirect(to string, args []interface{}) {
	router := GetModule(ModuleRouter).(*router)

	url := router.getReverseURL(to, args)
	c.Respw.Header().Set("Location", url)
	c.Respw.WriteHeader(302)
}

// Render gets the template for the calling action and renders it
func (c *Controller) Render() {
	tm := GetModule(ModuleTemplateManager).(*TemplateManager)

	if err := tm.RenderTemplate(c.Respw, c.Req, c.name+"/"+c.action+".html", c.RenderArgs); err != nil {
		Log.Warning(err.Error())
		http.Error(c.Respw, "500 internal server error", 500)
	}
}
