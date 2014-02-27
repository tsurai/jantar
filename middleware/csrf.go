package middleware

import(
  "os"
  "io"
  "fmt"
  "amber"
  "amber/context"
  "strings"
  "net/http"
  "html/template"
  "crypto/hmac"
  "crypto/sha512"
  "crypto/rand"
  "encoding/base64"
)

var (
  secretKey []byte
)

const (
  secretLength = 32
)

// Csrf is a Middleware that protects against cross-side request forgery
type Csrf struct {
  amber.Middleware
  // BlockingRandom determines whether to use /dev/urandom or /dev/random on unix systems
  BlockingRandom bool
}

func noAccess(respw http.ResponseWriter, req *http.Request) {
  http.Error(respw, "400 bad request", 400)
}

// Initialize prepares Csrf for usage
// Note: Do not call this yourself
func (c *Csrf) Initialize() {
  generateSecretKey(c.BlockingRandom)

  // add all hooks to TemplateManger
  tm := context.GetGlobal("TemplateManager").(*amber.TemplateManager)
  if tm == nil {
    panic("[Fatal] Failed to get template manager")
  }

  tm.AddTmplFunc("csrfToken", func() string { return "" })

  if err := tm.AddHook(amber.TmBeforeParse, beforeParseHook); err != nil {
    panic(err)
  }

  if err := tm.AddHook(amber.TmBeforeRender, beforeRenderHook); err != nil {
    panic(err)
  }
}

// Call executes the Middleware
// Note: Do not call this yourself
func (c *Csrf) Call(respw http.ResponseWriter, req *http.Request) bool {
  sessionID, foundSession := context.GetOk(req, "session_id")
  if !foundSession {
    return true
  }

  tokenString := req.PostFormValue("_csrf-token")
  if tokenString == "" {
    context.Set(req, "_csrf", base64.StdEncoding.EncodeToString(generateToken(sessionID.(string))))
    return true
  }

  if req.Method == "GET" || req.Method == "HEAD" {
    return true
  }

  token, _ := base64.StdEncoding.DecodeString(tokenString)
  if hmac.Equal(token, generateToken(sessionID.(string))) {
    return true
  }

  /* TODO: use error handler as parameter */
  noAccess(respw, req)
  fmt.Println("CSRF Detected! IP:", req.RemoteAddr)

  /* log ip etc pp */
  return false
}

func generateSecretKey(blocking bool) {
  secretKey = make([]byte, secretLength)
  
  if blocking {
    if file, err := os.Open("/dev/random"); err == nil {
      if n, err := io.ReadAtLeast(file, secretKey, secretLength); n == secretLength && err == nil {
        return
      }
    }
  }

  if n, err := rand.Read(secretKey); n != secretLength || err != nil {
    panic("[Fatal] Failed to generate secret key.")
  }
}

func generateToken(sessionID string) []byte {
  mac := hmac.New(sha512.New, secretKey)
  mac.Write([]byte(sessionID))
  
  return mac.Sum(nil)
}

func beforeParseHook(tm *amber.TemplateManager, name string, data *[]byte) {
  tmplData := string(*data)

  offset := strings.Index(tmplData, "<head>")
  if offset != -1 {
    tmplData = tmplData[:offset+6]+"<meta name=\"csrf-token\" content=\"{{csrfToken}}\">"+tmplData[offset+6:]
    *data = []byte(tmplData)
  }
}

func beforeRenderHook(req *http.Request, tm *amber.TemplateManager, tmpl *template.Template, args map[string]interface{}) {
  if token, ok := context.GetOk(req, "_csrf"); ok {
    tmpl = tmpl.Funcs(template.FuncMap{
      "csrfToken": func() string {
        return token.(string)
      },
    })
  }
}