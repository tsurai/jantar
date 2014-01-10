package amber

import (
  "fmt"
  "strings"
  "net/smtp"
)

type MailerConfig struct {
  Server  string
  Auth    smtp.Auth
}

type mailer struct {
  initialized bool
  client      *smtp.Client
  tm          *templateManager
  config      *MailerConfig
}

var Mailer mailer

func (m *mailer) initialize(tm *templateManager, config *MailerConfig) (err error) {
  if m.client, err = smtp.Dial(config.Server); err != nil {
    fmt.Println(err)
    return err
  }
  
  if ok, _ := m.client.Extension("STARTTLS"); ok {
    if err = m.client.StartTLS(nil); err != nil {
      return err
    }
  }
  
  if config.Auth != nil {
    if err = m.client.Auth(config.Auth); err != nil {
        return err
    }
  }

  m.tm = tm
  m.config = config
  m.initialized = true

  return nil
}

func (m *mailer) Send(from string, from_alias string, to []string, reply_to string, subject string, html bool, tmplname string, args map[string]string) (err error) {
  if !m.initialized {
    return fmt.Errorf("Mailer is not configured")
  }

  tmpl := m.tm.getTemplate("Mailer/" + tmplname + ".html")
  if tmpl == nil {
    return fmt.Errorf("Can't find template %s", tmplname)
  }

  headerData := make(map[string]string)
  headerData["From"] = from_alias + "<" + from + ">"
  headerData["To"] = strings.Join(to, ", ")
  headerData["Reply-To"] = reply_to
  headerData["Subject"] = subject
  headerData["MIME-Version"] = "1.0"
  if html {
    headerData["Content-Type"] = "text/html; charset=\"utf-8\""
  } else {
    headerData["Content-Type"] = "text/plain; charset=\"utf-8\""
  }

  header := ""
  for k, v := range headerData {
    if v != "" {
      header += fmt.Sprintf("%s: %s\r\n", k, v)
    }
  }

  if err = m.client.Mail(from); err != nil {
    return err
  }

  for _, addr := range to {
    if err = m.client.Rcpt(addr); err != nil {
      return err
    }
  }
  
  w, err := m.client.Data()
  if err != nil {
    return err
  }
  
  if _, err = w.Write([]byte(header)); err != nil {
    return err
  }

  if err = tmpl.Execute(w, args); err != nil {
    return fmt.Errorf("Failed to render template: %s", err.Error())
  }
  return w.Close()
}