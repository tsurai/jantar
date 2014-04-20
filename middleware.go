package jantar

import (
	"net/http"
	"reflect"
)

// Middleware implements core functionalities of the IMiddlware interface. Developer who want to write a Middleware
// should add Middleware as an anonymous field and implement Call().
type Middleware struct {
	next  *IMiddleware
	yield bool
}

// IMiddleware is an interface that describes a Middleware
type IMiddleware interface {
	Initialize()
	Cleanup()
	Call(rw http.ResponseWriter, r *http.Request) bool
	Yield(rw http.ResponseWriter, r *http.Request)
	setNext(mw *IMiddleware)
	doesYield() bool
}

func (m *Middleware) doesYield() bool {
	return m.yield
}

func (m *Middleware) setNext(mw *IMiddleware) {
	m.next = mw
}

// Yield suspends the current Middlware until all other Middlewares have been executed.
// This way a Middleware can execute code after all other Middlewares are done
func (m *Middleware) Yield(rw http.ResponseWriter, r *http.Request) {
	m.yield = true
	if m.next != nil {
		reflect.ValueOf(m.next).Elem().Interface().(IMiddleware).Call(rw, r)
	}
}
