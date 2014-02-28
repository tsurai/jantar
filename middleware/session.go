package middleware

import (
  "github.com/tsurai/amber"
  "github.com/tsurai/amber/context"
  "net/url"
  "net/http"
)

// Session is a Middleware stub for user sessions
type Session struct {
  amber.Middleware
  // CookieName is the name of the cookie saving the sessionID
  CookieName string
}

// Initialize prepares the Middleware for usage
func (s *Session) Initialize() {
  
}

// Call executes the Middleware
// Note: Do not call this yourself
func (s *Session) Call(respw http.ResponseWriter, req *http.Request) bool {
  // fetch flash from cookie
  if cookie, err := req.Cookie(s.CookieName); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      context.Set(req, "session_id", m["id"][0])
    }
  }
  
  return true
}