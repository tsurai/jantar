package amber

import (
  "errors"
  "reflect"
)

var (
  HookDuplicateId     = errors.New("Id already in use")
  HookUnknownId       = errors.New("Unknown hook Id")
  HookInvalidHandler  = errors.New("Handler is not a function")
)

type hooks struct {
  ids     []int
  handler map[int][]interface{}
}

func (h *hooks) registerHookId(hookId int) error {
  if h.handler == nil {
    h.handler = make(map[int][]interface{})
  }

  if h.isKnownId(hookId) {
    return HookDuplicateId
  }

  h.ids = append(h.ids, hookId)

  return nil
}

func (h *hooks) getHooks(hookId int) []interface{} {
  handler, ok := h.handler[hookId]
  if !ok {
    return nil
  }

  return handler
}

func (h *hooks) AddHook(hookId int, handler interface{}) error {
  if !h.isKnownId(hookId) {
    return HookUnknownId
  }

  if reflect.TypeOf(handler).Kind() != reflect.Func {
    return HookInvalidHandler
  }

  h.handler[hookId] = append(h.handler[hookId], handler)
  return nil
}

func (h *hooks) isKnownId(hookId int) bool {
  for _, id := range h.ids {
    if id == hookId {
      return true
    }
  }

  return false
}