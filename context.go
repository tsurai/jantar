package amber

import (
	"reflect"
	"net/http"
)

type context struct {
	handler Handler
	route		*route
	tm 			*tmplManager
	rw 			http.ResponseWriter
	req 		*http.Request
	payload map[reflect.Type]reflect.Value
}

func newContext(h Handler, route *route, rw http.ResponseWriter, req *http.Request, tm *tmplManager, params Param) *context {
	ctx := &context{handler: h, route: route, rw: rw, req: req, tm: tm, payload: make(map[reflect.Type]reflect.Value)}
	ctx.addPayload(rw)
	ctx.addPayload(req)
	ctx.addPayload(params)

	return ctx
}

func (ctx *context) callHandler() {
	t := reflect.TypeOf(ctx.handler)
	var in []reflect.Value

	i := 0
	if ctx.route.isController {
		i = 1
		c := newController(ctx)
		in = append(in, reflect.ValueOf(c))
	}

	for ; i < t.NumIn(); i++ {
		if t.In(i).Kind() == reflect.Interface {
			for k, v := range ctx.payload {
				if k.Implements(t.In(i)) {
					in = append(in, v)
				}
			}
		} else if v := ctx.payload[t.In(i)]; v.IsValid() {
			in = append(in, v)
		}
	}

	if len(in) != t.NumIn() {
		logger.Printf("Invalid parameter count! Expected %d but got %d\n", t.NumIn(), len(in))
	}

	// TODO: catch exception
	ret := reflect.ValueOf(ctx.handler).Call(in)
	if len(ret) > 0 {
		ctx.rw.Write([]byte(ret[0].String()))
	}
}

func (ctx *context) addPayload(v interface{}) {
	ctx.payload[reflect.TypeOf(v)] = reflect.ValueOf(v)
}