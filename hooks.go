package jantar

import (
  "errors"
  "reflect"
)

// TODO: add hook fire function

var (
  ErrHookDuplicateID = errors.New("id already in use")
  ErrHookUnknownID = errors.New("unknown hook Id")
  ErrHookInvalidHandler = errors.New("handler is not a function")
)

type hook struct {
  signiture reflect.Type
  handler []interface{}
}

type hooks struct {
  list map[int]*hook
}

func (h *hooks) registerHook(hookID int, signiture reflect.Type) error {
  if h.list == nil {
    h.list = make(map[int]*hook)
  }

  if _, ok := h.list[hookID]; ok {
    Log.Errord(JLData{"id": hookID}, "Failed to register hook: id already in use")
    return ErrHookDuplicateID
  }

  if signiture.Kind() != reflect.Func {
    Log.Errord(JLData{"signiture": signiture}, "Failed to register hook: signiture is not a function")
    return ErrHookInvalidHandler
  }

  h.list[hookID] = &hook{signiture, nil}

  return nil
}

func (h *hooks) getHooks(hookID int) []interface{} {
  hook, ok := h.list[hookID]
  if !ok {
    Log.Errord(JLData{"id": hookID}, "Failed to get hook: unknown id")
    return nil
  }

  return hook.handler
}

func (h *hooks) AddHook(hookID int, handler interface{}) error {
  hook, ok := h.list[hookID]
  if !ok {
    Log.Errord(JLData{"id": hookID}, "Failed to add hook: unknown id")
    return ErrHookUnknownID
  }

  if !reflect.TypeOf(handler).AssignableTo(hook.signiture) {
    Log.Errord(JLData{"given": reflect.TypeOf(handler), "wanted": hook.signiture}, "Failed to add hook: handler type doesn't match the signiture")
    return ErrHookInvalidHandler
  }

  hook.handler = append(hook.handler, handler)
  return nil
}