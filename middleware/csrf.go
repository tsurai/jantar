package middleware

import(
  "amber"
  "amber/context"
  "regexp"
  "net/http"
  "html/template"
  "crypto/hmac"
  "crypto/sha512"
  "crypto/rand"
  "encoding/base64"
)

var (
  secretkey []byte
)

const (
  field_name = "csrf"
  secret_length = 32
)

type Csrf struct {
  amber.Middleware
}

func (c *Csrf) Initialize() {
  generateSecretKey()

  // add all hooks to TemplateManger
  tm := context.GetGlobal("TemplateManager").(*amber.TemplateManager)
  if tm == nil {
    panic("[Fatal] Failed to get template manager")
  }

  if err := tm.AddHook(amber.TM_BEFORE_PARSE, beforeParseHook); err != nil {
    panic(err)
  }

  if err := tm.AddHook(amber.TM_BEFORE_RENDER, beforeRenderHook); err != nil {
    panic(err)
  }
}

func (c *Csrf) Name() string {
  return "Csrf"
}

func (c *Csrf) Call(rw http.ResponseWriter, r *http.Request) {
  sessionId, foundSession := context.GetOk(r, "session_id")
  if !foundSession {
    // proceed to handler
    return
  }
 
  tokenString := r.PostFormValue(field_name)
  if tokenString == "" {
    context.Set(r, field_name, base64.StdEncoding.EncodeToString(generateToken(sessionId.(string))))
  }

  if r.Method == "GET" || r.Method == "HEAD" {
    return
  }

  /*if r.URL.Scheme == "https" {
    do same origin check
  }*/

  token, _ := base64.StdEncoding.DecodeString(tokenString)
  if hmac.Equal(token, generateToken(sessionId.(string))) {
    // proceed to handler
    return
  }

  // show error
  rw.WriteHeader(400)
  rw.Write([]byte("CSRF"))
}

func generateSecretKey() {
  secretkey := make([]byte, secret_length)
  
/*
  if file, err := os.Open("/dev/random"); err == nil {
    if n, err := io.ReadAtLeast(file, secretkey, secret_length); n == secret_length && err == nil {
      return
    }
  }
*/

  if n, err := rand.Read(secretkey); n != secret_length || err != nil {
    panic("[Fatal] Failed to generate secret key.")
  }
}

func generateToken(sessionId string) []byte {
  mac := hmac.New(sha512.New, secretkey)
  mac.Write([]byte(sessionId))

  return mac.Sum(nil)
}

func beforeParseHook(tm *amber.TemplateManager, name string, data *[]byte) {
  tmplData := string(*data)

  regex := regexp.MustCompile("(?i:<form .* method=(\"|')(POST|PUT|DELETE)(\"|').*>)")
  tmplData = regex.ReplaceAllStringFunc(tmplData, func(match string) string {
    return match+"<input type='hidden' name='" + field_name + "' value='{{" + field_name + "}}'>"
  })

  *data = []byte(tmplData)
}

func beforeRenderHook(tm *amber.TemplateManager, tmpl *template.Template, args map[string]interface{}) {
  tmpl = tmpl.Funcs(template.FuncMap{
    "csrf": func() string {
      return args[field_name].(string)
    },
  })
}