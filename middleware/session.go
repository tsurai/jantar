package middleware

import (
  "amber"
  "amber/context"
  "net/url"
  "net/http"
)

type Session struct {
  amber.Middleware
  CookieName string
}

func (s *Session) Initialize() {
  
}

func (s *Session) Name() string {
  return "Session"
}

func (s *Session) Call(rw http.ResponseWriter, r *http.Request) {
  // fetch flash from cookie
  if cookie, err := r.Cookie(s.CookieName); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      context.Set(r, "session_id", m["id"][0])
    }
  }
}