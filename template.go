package jantar

import (
  "github.com/tsurai/jantar/context"
  "os"
  "fmt"
  "time"
  "strings"
  "strconv"
  "net/http"
  "io/ioutil"
  "path/filepath"
  "html/template"
  "github.com/howeyc/fsnotify"
)

const (
  // TmBeforeParse is a hook with callback f(tm *TemplateManager, name string, data string)
  TmBeforeParse = iota
  // TmBeforeRender is a hook with callback f(req *http.Request, tm *TemplateManager, tmpl *html.Template, args map[string]interface{})
  TmBeforeRender = iota
  tmLast = iota
)

// TemplateManager is responsible for loading, watching and rendering templates
type TemplateManager struct {
  hooks
  directory     string
  watcher       *fsnotify.Watcher
  tmplFuncs     template.FuncMap
  tmplList      *template.Template
}

func newTemplateManager(directory string) *TemplateManager {
  funcs := template.FuncMap{
    /* security functions */
    "antiClickjacking": func() template.HTML {
      return  template.HTML("<style id=\"antiClickjack\">body{display:none !important;}</style>")
    },
    "set": func(args map[string]interface{}, key string, value interface{}) string {
      if args != nil {
        args[key] = value
      }
      return ""
    },
    "array": func(args ...interface{}) []interface{} {
      var ret []interface{}
      for _, arg := range args {
        ret = append(ret, arg)
      }
      return ret
    },
    "errorClass": func(errors []string) string {
      if errors != nil {
          return "has-error"
      }
      return ""
    },
    "toHtml": func(str string) template.HTML {
      return template.HTML(str)
    },
    "url": func(name string, args ...interface{}) string {
      router := context.GetGlobal("Router").(*router)
      return router.getReverseUrl(name, args)
    },/*
    "flash": func(args map[string]interface{}, key string) string {
      return args["flash"].(map[string]string)[key]
    },*/
    "since": func(t time.Time) string {
      seconds := int(time.Since(t).Seconds())
      if seconds < 60 {
        return "< 1 minute ago"
      } else if seconds < 60 * 2 {
        return "1 minute ago"
      } else if seconds < 60 * 60 {
        return fmt.Sprintf("%d minutes ago", seconds / 60)
      } else if seconds < 60 * 60 * 2 {
        return "1 hour ago"
      } else if seconds < 60 * 60 * 24 {
        return fmt.Sprintf("%d hours ago", seconds / (60 * 60))
      } else if seconds < 60 * 60 * 24 * 2 {
        return "1 day ago"
      } else if seconds < 60 * 60 * 24 * 30 {
        return fmt.Sprintf("%d days ago", seconds / (60 * 60 * 24))
      } else if seconds < 60 * 60 * 24 * 30 * 2 {
        return "1 month ago"
      } else if seconds < 60 * 60 * 24 * 30 * 12 {
        return fmt.Sprintf("%d months ago", seconds / (60 * 60 * 24 * 30))
      }
      return "> 1 year ago"
    },
    "paginate": func(curPage int, nPages int, offset int, url string) template.HTML {
      if nPages < 2 {
        return template.HTML("")
      }

      result := "<ul class='pagination'>"

      if curPage > 1 {
        result += "<li><a href='" + url + "/page/first'>&laquo;First</a></li>" +
                  "<li><a href='" + url + "/page/" + strconv.Itoa(curPage - 1) + "'>&laquo;</a></li>"
      }

      if curPage - offset > 1 {
        result += "<li><span>...</span></li>"
      }

      for i := curPage - offset; i < curPage + offset + 1; i++ {
        if i > 0 && i <= nPages {
          if i == curPage {
            result += "<li class='active'><a href='" + url + "/page/" + strconv.Itoa(i) + "'>" + strconv.Itoa(i) + "</a></li>"
          } else {
            result += "<li><a href='" + url + "/page/" + strconv.Itoa(i) + "'>" + strconv.Itoa(i) + "</a></li>"    
          }
        }
      }

      if curPage + offset < nPages {
        result += "<li><span>...</span></li>"
      }

      if curPage != nPages {
        result += "<li><a href='" + url + "/page/" + strconv.Itoa(curPage + 1) + "'>&raquo;</a></li>" +
                  "<li><a href='" + url + "/page/last'>Last&raquo;</a></li>"
      }
      return template.HTML(result + "</ul>")
    },
  }

  tm := &TemplateManager{directory: directory, tmplFuncs: funcs}
  
  // register hooks
  for i := 0; i < tmLast; i++ {
    if err := tm.registerHookID(i); err != nil {
      panic("Failed to register hook." + err.Error())
    }
  }

  return tm
}

// watch listens for file events and reloads templates on changes
func (tm *TemplateManager) watch() {
  for {
    select {
    case ev := <- tm.watcher.Event:
      if !ev.IsRename() && filepath.Ext(ev.Name) == ".html" {
        logger.Println("Reloading templates...")
        tm.loadTemplates()
        return
      }
    case err := <-tm.watcher.Error:
      logger.Println("[Warning] File Watcher:", err)
      return
    }
  } 
}

func (tm *TemplateManager) loadTemplates() error {
  var err error
  var templates *template.Template

  // close watcher if running
  if tm.watcher != nil {
    tm.watcher.Close()
  }
  
  // create a new watcher and start the watcher thread
  if tm.watcher, err = fsnotify.NewWatcher(); err != nil {
    return fmt.Errorf("[Warning] Can't create new fswatcher. Reason: %s", err.Error())
  }
  go tm.watch()

  // walk resursive through the template directory
  res := filepath.Walk(tm.directory, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return fmt.Errorf("[Warning] Can't walk directory '%s'. Reason: %s", path, err.Error())
    }

    if info.IsDir() {
      if strings.HasPrefix(info.Name(), ".") {
        return filepath.SkipDir
      }

      // add the current directory to the watcher
      if err = tm.watcher.Watch(path); err != nil {
        logger.Printf("[Warning] Can't watch directory '%s'. Reason: %s\n", path, err.Error())
      }
      return nil
    }

    if strings.HasSuffix(info.Name(), ".html") {
      fdata, err := ioutil.ReadFile(path)
      if err != nil {
        return fmt.Errorf("[Warning] Can't read template file '%s'. Reason: %s\n", err.Error())
      }

      tmplName := strings.Replace(strings.ToLower(path[len(tm.directory)+1:]), "\\", "/", -1)
      
      // call BEFORE_PARSE hooks
      hooks := tm.getHooks(TmBeforeParse)
      for _, hook := range hooks {
        hook.(func(*TemplateManager, string, *[]byte))(tm, tmplName, &fdata)
      }
      
      // add the custom template functions to the first template
      if templates == nil {
        templates, err = template.New(tmplName).Funcs(tm.tmplFuncs).Parse(string(fdata))
      } else {
        _, err = templates.New(tmplName).Parse(string(fdata))
      }

      if err != nil {
        return fmt.Errorf("[Warning] Failed to parse template '%s'. Reason: %s", tmplName, err.Error())
      }
    }
    return nil
  })

  // no errors occured, override the old list
  if res == nil {
    tm.tmplList = templates
  }

  return res
}

func (tm *TemplateManager) getTemplate(name string) *template.Template {
  if tm.tmplList == nil {
    return nil
  }

  return tm.tmplList.Lookup(strings.ToLower(name))
}

// AddTmplFunc adds a template function with a given name and function pointer.
// Note: AddTmplFunc has no effect if called after the templates have been parsed.
func (tm *TemplateManager) AddTmplFunc(name string, fn interface{}) {
  tm.tmplFuncs[name] = fn
}

// RenderTemplate renders a template with the given name and arguments.
// Note: A Controller should call its Render function instead.
func (tm *TemplateManager) RenderTemplate(respw http.ResponseWriter, req *http.Request, name string, args map[string]interface{}) error {
  tmpl := tm.tmplList.Lookup(strings.ToLower(name))
  if tmpl == nil {
    return fmt.Errorf("[Warning] Can't find template '%s'", strings.ToLower(name))
  }

  // call BEFORE_RENDER hooks
  hooks := tm.getHooks(TmBeforeRender)
  for _, hook := range hooks {
    hook.(func(*http.Request, *TemplateManager, *template.Template, map[string]interface{}))(req, tm, tmpl, args)
  }

  if err := tmpl.Execute(respw, args); err != nil {
    return fmt.Errorf("[Warning] Failed to render template. Reason: %s", err.Error())
  }

  return nil
}