package context

import (
  "net/http"
)

var (
  global_data = make(map[interface{}]interface{})
  request_data = make(map[*http.Request]map[interface{}]interface{})
)

func SetGlobal(key, val interface{}) {
  global_data[key] = val
}

func GetGlobal(key interface{}) interface{} {
  return global_data[key]
}

func GetGlobalOk(key interface{}) (interface{}, bool) {
  val, ok := global_data[key]
  return val, ok
}

func Set(r *http.Request, key, val interface{}) {
  if request_data[r] == nil {
    request_data[r] = make(map[interface{}]interface{})
  }
  request_data[r][key] = val
}

func Get(r *http.Request, key interface{}) interface{} {
  if request_data[r] != nil {
    return request_data[r][key]
  } else {
    return nil
  }
}

func GetOk(r *http.Request, key interface{}) (interface{}, bool) {
  if request_data[r] == nil {
    return nil, false
  }
  val, ok := request_data[r][key]
  return val, ok
}