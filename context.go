package amber

import (
	"reflect"
	"net/http"
)

type context struct {
	handler Handler
	route		*route
	tm 			*templateManager
	rw 			http.ResponseWriter
	Req 		*http.Request
	payload map[reflect.Type]reflect.Value
}

func newContext(h Handler, route *route, rw http.ResponseWriter, req *http.Request, tm *templateManager, params Param) *context {
	ctx := &context{h, route, tm, rw, req, make(map[reflect.Type]reflect.Value)}

	if req.Method == "POST" {
			req.ParseForm()
	}

	// TODO: find a better solution
	ctx.addPayload(rw)
	ctx.addPayload(req)
	ctx.addPayload(params)

	return ctx
}

func (ctx *context) callHandler() {
	handlerType := reflect.TypeOf(ctx.handler)
	var in []reflect.Value

	i := 0
	if ctx.route.isController {
		i = 1
		c := newController(ctx)
		in = append(in, reflect.ValueOf(c))
	}

	for ; i < handlerType.NumIn(); i++ {
		paramType := handlerType.In(i)
		if paramType.Kind() == reflect.Interface {
			for k, v := range ctx.payload {
				if k.Implements(paramType) {
					in = append(in, v)
				}
			}
		} else if v := ctx.payload[paramType]; v.IsValid() {
			in = append(in, v)
		} else if (ctx.route.method == "POST") && (paramType.Kind() == reflect.Ptr) && (paramType.Elem().Kind() == reflect.Struct) {
			if v := ParsePostData(ctx.Req.PostForm, paramType); !v.IsNil() {
				in = append(in, v)
			}
		}
	}

	if len(in) != handlerType.NumIn() {
		logger.Printf("![Warning]! Invalid parameter count! Expected %d but got %d\n", handlerType.NumIn(), len(in))
		return
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