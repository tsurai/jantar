package jantar

import (
  "errors"
  "reflect"
)

var (
  ErrHookDuplicateID     = errors.New("id already in use")
  ErrHookUnknownID       = errors.New("unknown hook Id")
  ErrHookInvalidHandler  = errors.New("handler is not a function")
)

type hooks struct {
  ids     []int
  handler map[int][]interface{}
}

func (h *hooks) registerHookID(hookID int) error {
  if h.handler == nil {
    h.handler = make(map[int][]interface{})
  }

  if h.isKnownID(hookID) {
    return ErrHookDuplicateID
  }

  h.ids = append(h.ids, hookID)

  return nil
}

func (h *hooks) getHooks(hookID int) []interface{} {
  handler, ok := h.handler[hookID]
  if !ok {
    return nil
  }

  return handler
}

func (h *hooks) AddHook(hookID int, handler interface{}) error {
  if !h.isKnownID(hookID) {
    return ErrHookUnknownID
  }

  if reflect.TypeOf(handler).Kind() != reflect.Func {
    return ErrHookInvalidHandler
  }

  h.handler[hookID] = append(h.handler[hookID], handler)
  return nil
}

func (h *hooks) isKnownID(hookID int) bool {
  for _, id := range h.ids {
    if id == hookID {
      return true
    }
  }

  return false
}