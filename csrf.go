package jantar

import(
  "github.com/tsurai/jantar/context"
  "strings"
  "net/url"
  "net/http"
  "html/template"
  "crypto/rand"
  "crypto/hmac"
  "crypto/sha512"
  "encoding/base64"
)

// TODO: accept custom handler

var (
  secretKey []byte
)

const (
  secretLength = 32
)

// csrf is a Middleware that protects against cross-side request forgery
type csrf struct {
  Middleware
}

func noAccess(respw http.ResponseWriter, req *http.Request) {
  http.Error(respw, "400 bad request", 400)
}

// Initialize prepares csrf for usage
// Note: Do not call this yourself
func (c *csrf) Initialize() {
  generateSecretKey()

  // add all hooks to TemplateManger
  tm := context.GetGlobal("TemplateManager").(*TemplateManager)
  if tm == nil {
    logger.Fatal("Failed to get template manager")
  }

  tm.AddTmplFunc("csrfToken", func() string { return "" })

  if err := tm.AddHook(TmBeforeParse, beforeParseHook); err != nil {
    logger.Fatal(err)
  }

  if err := tm.AddHook(TmBeforeRender, beforeRenderHook); err != nil {
    logger.Fatal(err)
  }
}

// Cleanup saves the current secretkey to accept old tokens with the next start
func (c *csrf) Cleanup() {
  // TODO: Save last secretkey for the next start
}

// Call executes the Middleware
// Note: Do not call this yourself
func (c *csrf) Call(respw http.ResponseWriter, req *http.Request) bool {
  uniqueID := make([]byte, 32)
  
  if cookie, err := req.Cookie("JANTAR_ID"); err == nil {
    if m, err := url.ParseQuery(cookie.Value); err == nil {
      uniqueID, _ = base64.StdEncoding.DecodeString(m["id"][0])
    }
  } else {
    if n, err := rand.Read(uniqueID); n != 32 || err != nil {
      logger.Fatal("Failed to generate secret key")
    }

    http.SetCookie(respw, &http.Cookie{Name: "JANTAR_ID", Value: "id=" + base64.StdEncoding.EncodeToString(uniqueID)})
  }

  context.Set(req, "_csrf", base64.StdEncoding.EncodeToString(generateToken(uniqueID)))

  if req.Method == "GET" || req.Method == "HEAD" {
    return true
  }

  token, _ := base64.StdEncoding.DecodeString(req.PostFormValue("_csrf-token"))
  if hmac.Equal(token, generateToken(uniqueID)) {
    return true
  }

  /* TODO: use error handler as parameter */
  noAccess(respw, req)
  logger.Errord(JLData{"IP": req.RemoteAddr}, "CSRF detected!")

  /* log ip etc pp */
  return false
}

func generateSecretKey() {
  secretKey = make([]byte, secretLength)

  if n, err := rand.Read(secretKey); n != secretLength || err != nil {
    logger.Fatal("Failed to generate secret key")
  }
}

func generateToken(uniqueID []byte) []byte {
  mac := hmac.New(sha512.New, secretKey)
  mac.Write(uniqueID)
  
  return mac.Sum(nil)
}

func beforeParseHook(tm *TemplateManager, name string, data *[]byte) {
  tmplData := string(*data)

  offset := strings.Index(tmplData, "<head>")
  if offset != -1 {
    tmplData = tmplData[:offset+6] + "<meta name=\"csrf-token\" content=\"{{csrfToken}}\">" + tmplData[offset+6:]
    *data = []byte(tmplData)
  }
}

func beforeRenderHook(req *http.Request, tm *TemplateManager, tmpl *template.Template, args map[string]interface{}) {
  if token, ok := context.GetOk(req, "_csrf"); ok {
    tmpl = tmpl.Funcs(template.FuncMap{
      "csrfToken": func() string {
        return token.(string)
      },
    })
  }
}