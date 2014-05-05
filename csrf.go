package jantar

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/tsurai/jantar/context"
	"html/template"
	"net/http"
	"strings"
)

// TODO: accept custom handler

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
	// add all hooks to TemplateManger
	tm := GetModule(ModuleTemplateManager).(*TemplateManager)
	if tm == nil {
		Log.Fatal("failed to get template manager")
	}

	tm.AddTmplFunc("csrfToken", func() string { return "" })

	tm.AddHook(TmBeforeParse, beforeParseHook)
	tm.AddHook(TmBeforeRender, beforeRenderHook)
}

// Cleanup saves the current secretkey to accept old tokens with the next start
func (c *csrf) Cleanup() {
	// TODO: Save last secretkey for the next start
}

// Call executes the Middleware
// Note: Do not call this yourself
func (c *csrf) Call(respw http.ResponseWriter, req *http.Request) bool {
	var cookieToken []byte

	if cookie, err := req.Cookie("JANTAR_ID"); err == nil {
		cookieToken, _ = hex.DecodeString(cookie.Value)
	} else {
		cookieToken = make([]byte, 32)
		if n, err := rand.Read(cookieToken); n != 32 || err != nil {
			Log.Fatal("failed to generate secret key")
		}

		http.SetCookie(respw, &http.Cookie{Name: "JANTAR_ID", Value: hex.EncodeToString(cookieToken)})
	}

	context.Set(req, "_csrf", cookieToken, true)

	// check for safe methods
	if req.Method == "GET" || req.Method == "HEAD" {
		return true
	}

	if req.PostFormValue("_csrf-token") == string(cookieToken) {
		return true
	}

	/* TODO: use error handler as parameter */
	noAccess(respw, req)
	Log.Errord(JLData{"IP": req.RemoteAddr}, "CSRF detected!")

	/* log ip etc pp */
	return false
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
