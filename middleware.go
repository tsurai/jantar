package amber

import (
  "reflect"
  "net/http"
)

type Middleware struct {
  next *IMiddleware
  yield bool
}

type IMiddleware interface {
  Initialize()
  Name() string
  Call(rw http.ResponseWriter, r *http.Request)
  Next(rw http.ResponseWriter, r *http.Request)
  setNext(mw *IMiddleware)
  doesYield() bool
}

func (m *Middleware) doesYield() bool {
  return m.yield
}

func (m *Middleware) setNext(mw *IMiddleware) {
  m.next = mw
}

func (m *Middleware) Next(rw http.ResponseWriter, r *http.Request) {
  m.yield = true
  if m.next != nil {
    reflect.ValueOf(m.next).Elem().Interface().(IMiddleware).Call(rw, r)
  }
}