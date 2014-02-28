// Context is a small package for arbitrary per-request and global data.
//
// For now this basically is a copy of Gorilla/Context but it will be more customized in future versions.

package context

import (
  "net/http"
)

var (
  globalData = make(map[interface{}]interface{})
  requestData = make(map[*http.Request]map[interface{}]interface{})
)

// SetGlobal saves a value with given key in the global context
func SetGlobal(key, val interface{}) {
  globalData[key] = val
}

// GetGlobal searches for a value with given key in the global context and returns it
func GetGlobal(key interface{}) interface{} {
  return globalData[key]
}

// GetGlobalOk does the same as GetGlobal but returns an additional boolean indicating if a value with the given key was found
func GetGlobalOk(key interface{}) (interface{}, bool) {
  val, ok := globalData[key]
  return val, ok
}

// Set saves the a value with given key for a specific http.Request
func Set(req *http.Request, key, val interface{}) {
  if requestData[req] == nil {
    requestData[req] = make(map[interface{}]interface{})
  }
  requestData[req][key] = val
}

// Get returns a value with given name and request
func Get(req *http.Request, key interface{}) interface{} {
  if requestData[req] != nil {
    return requestData[req][key]
  }
  return nil
}

// GetOk does the same as Get but returns an additional boolean indicating if a value with the given key and request was found
func GetOk(req *http.Request, key interface{}) (interface{}, bool) {
  if requestData[req] == nil {
    return nil, false
  }
  val, ok := requestData[req][key]
  return val, ok
}

func ClearData(req *http.Request) {
  delete(requestData, req)
}