// Context is a small package for arbitrary per-request, global data and modules
//
// Module data are set by jantar and are read-only by default.
package context

import (
  "net/http"
)

// TODO: add special read-only module data and add read-only mode to the normal data stores as well

type data struct {
  value     interface{}
  readOnly  bool
}

var (
  globalData = make(map[interface{}]data)
  requestData = make(map[*http.Request]map[interface{}]data)
)

// SetGlobal saves a value with given key in the global context
func SetGlobal(key, value interface{}, readOnly bool) {
  if gd, ok := globalData[key]; !ok || !gd.readOnly {
    globalData[key] = data{value, readOnly}
  }
}

// GetGlobal searches for a value with given key in the global context and returns it
func GetGlobal(key interface{}) interface{} {
  return globalData[key].value
}

// GetGlobalOk does the same as GetGlobal but returns an additional boolean indicating if a value with the given key was found
func GetGlobalOk(key interface{}) (interface{}, bool) {
  gd, ok := globalData[key]
  return gd.value, ok
}

// Set saves the a value with given key for a specific http.Request
func Set(req *http.Request, key, value interface{}, readOnly bool) {
  rd, ok := requestData[req]
  if !ok && rd == nil {
    requestData[req] = make(map[interface{}]data)
  }

  if d, ok := rd[key]; !ok || !d.readOnly {
    requestData[req][key] = data{value, readOnly}
  }
}

// Get returns a value with given name and request
func Get(req *http.Request, key interface{}) interface{} {
  if requestData[req] != nil {
    return requestData[req][key].value
  }
  return nil
}

// GetOk does the same as Get but returns an additional boolean indicating if a value with the given key and request was found
func GetOk(req *http.Request, key interface{}) (interface{}, bool) {
  if requestData[req] == nil {
    return nil, false
  }
  
  d, ok := requestData[req][key]
  return d.value, ok
}

func ClearData(req *http.Request) {
  delete(requestData, req)
}