package amber

import (
  "strings"
  "os"
  "io/ioutil"
  "path/filepath"
  "html/template"
  "github.com/howeyc/fsnotify"
)

type templateManager struct {
  directory     string
  router        *router
  watcher       *fsnotify.Watcher
  tmplFuncs     template.FuncMap
  tmplList      *template.Template
}

func newtemplateManager(directory string, router *router) *templateManager {
  funcs := template.FuncMap{
    "set": func(args map[string]interface{}, key string, value interface{}) string {
      args[key] = value
      return ""
    },
    "errorClass": func(errors []string) string {
      if errors != nil {
          return "has-error"
      }
      return ""
    },
    "url": func(name string, args ...interface{}) string {
      return router.getReverseUrl(name, args)
    },
  }

  return &templateManager{directory: directory, router: router, tmplFuncs: funcs}
}

// watcher listens for file events
func (tm *templateManager) watch() {
  for {
    select {
    case ev := <- tm.watcher.Event:
      if !ev.IsRename() && filepath.Ext(ev.Name) == ".html" {
        logger.Println("Reloading templates...")
        tm.loadTemplates()
        return
      }
    case err := <-tm.watcher.Error:
      logger.Println("![Warning]! File Watcher:", err)
      return
    }
  } 
}

func (tm *templateManager) loadTemplates() {
  var err error
  var templates *template.Template

  // close watcher if running
  if tm.watcher != nil {
    tm.watcher.Close()
  }
  
  // create a new watcher and start the watcher thread
  if tm.watcher, err = fsnotify.NewWatcher(); err != nil {
    panic("Failed to create new watcher:" + err.Error())
  }
  go tm.watch()

  // walk resursive through the template directory
  filepath.Walk(tm.directory, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      panic(err)
    }

    if info.IsDir() {
      if strings.HasPrefix(info.Name(), ".") {
        return filepath.SkipDir
      }

      // add the current directory to the watcher
      if err = tm.watcher.Watch(path); err != nil {
        logger.Println("![Warning]! Can't watch directory %s. %s", path, err.Error())
      }
      return nil
    }

    if strings.HasSuffix(info.Name(), ".html") {
      fdata, err := ioutil.ReadFile(path)
      if err != nil {
        logger.Println("![Warning]! Failed to read template file", path)
        return nil
      }

      tmplName := strings.Replace(strings.ToLower(path[len(tm.directory)+1:]), "\\", "/", -1)

      // add the custom template functions to the first template
      if templates == nil {
        templates, err = template.New(tmplName).Funcs(tm.tmplFuncs).Parse(string(fdata))
      } else {
        _, err = templates.New(tmplName).Parse(string(fdata))
      }

      if err != nil {
        logger.Println("![Warning]! Failed to parse template " + tmplName + ". " + err.Error())
        return nil
      }
    }
    return nil
  })

  // no errors occured, override the old list
  tm.tmplList = templates
}

func (tm *templateManager) getTemplate(name string) *template.Template {
  if tm.tmplList == nil {
    return nil
  }
  
  return tm.tmplList.Lookup(strings.ToLower(name))
}