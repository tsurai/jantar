# Jantar [![GoDoc](https://godoc.org/github.com/tsurai/jantar?status.png)](http://godoc.org/github.com/tsurai/jantar)

Jantar is a lightweight mvc web framework with emphasis on security written in golang. It has been largely inspired by [Martini](https://github.com/codegangsta/martini) but prefers performance over syntactic sugar and aims to provide crucial security settings and features right out of the box.

## Features
* RESTful pattern with protection against cross-site request forgery
* Secure default settings for TLS
	* No RC4, DES or similarl insecure cipher
	* No SSL, requires at least TLS 1.0
	* Prefered cipher: TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
* Secure default HTTP header
	* Strict-Transport-Security: max-age=31536000;includeSubDomains
	* X-Frame-Options: sameorigin
	* X-XSS-Protection: 1;mode=block
	* X-Content-Type-Options: nosniff
* Encrypted and signed cookies using AES256 and HMAC(SHA256)
* Simple Middleware interface
* Compatible with http.HandlerFunc
* Responsive to current events
	* ![Preview](https://i.imgur.com/OKGR3WG.png)

## Table of Contents
* [Current State](#current-state)
* [Getting Started](#getting-started)
  * [Controller](#controller)
* [A note on security](#a-note-on-security)
	* [/dev/urandom](#devurandom)
* [Todo List](#todo-list)

## Current State
Jantar is currently getting completly redesigned and is not usable right now.

## Getting Started

First you have to download and install the package into the import path. This can easily be done with go get:
```
go get github.com/tsurai/jantar
```

Now you can import jantar and create a simple website
```go
package main

import (
	"net/http"
	"github.com/tsurai/jantar"
)

func main() {
	j := jantar.New(&jantar.Config {
		Hostname: "localhost",
		Port:     3000,
	})

	j.AddRoute("GET", "/", func(respw http.ResponseWriter, req *http.Request) {
		respw.Write([]byte("Hello World"))
	})

	j.Run()
}
```

### Controller

Using Controller and rendering Templates is very easy with Jantar. For this simple example I'm going to assume the following directory structure. A detailed description will follow soon.
```
|- controllers/
|-- app.go
|- views/
|-- app/
|--- index.html
| main.go
```
--
*controllers/app.go*
```go
package controller

import (
	"github.com/tsurai/jantar"
)

type App struct {
	jantar.Controller
}

func (c *App) Index() {
	c.Render()
}
```
--
*views/app/index.html*
```html
<h1>Hello Controller</h1>
```
--
*main.go*
```go
package main

import (
	"github.com/tsurai/jantar"
	c "controllers"
)

func main() {
	j := jantar.New(&jantar.Config {
		Hostname: "localhost",
		Port:     3000,
	})

	j.AddRoute("GET", "/", (*c.App).Index)

	j.Run()
}

```

## A note on security
Jantar is by no means secure in the literal sense of the word. What it does is providing easy and fast ways to protect against the most common vulnerabilities. Security should never be left out because it is too troublesome to implement.

### /dev/urandom
Some might wonder why Jantar is using /dev/urandom instead of the seemingly more secure /dev/random.
Please take some minutes and read this interesting article about [/dev/urandom/](http://www.2uo.de/myths-about-urandom/)

## Todo List
* more consistency in error handling
* flexible config
	* custom public directory
	* custom http response handler
	* custom cookie names for internal modules
* completion of controller struct
* reintroduction of models
* more hooks and a better middleware interface

## Inspirations
These frameworks have been a source of inspiration and ideas during Jantars development.
* [Martini](https://github.com/go-martini/martini)
* [Web](https://github.com/gocraft/web)
* [Revel](https://github.com/revel/revel)
