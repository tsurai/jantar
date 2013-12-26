# Amber

Amber is a very minimalist mcv web framework written in golang. It has been largely inspired by [Martini](https://github.com/codegangsta/martini) and [Revel](https://github.com/robfig/revel).

## Getting Started

First you have to download and install the package into the import path. This can easily be done with go get:
```
go get github.com/tsurai/amber
```

Now you can import amber and create a simple website
```
package main

import (
	"github.com/tsurai/amber"
)

func main() {
	a := amber.New()

	a.AddRoute("GET", "/, func() string {
		return "Hello World"
	})

	a.Run()
}
```

## Controller

Using Controller and rendering Templates is very easy with Amber. For this simple example I'm going to assume the following directory structure. A detailed description will follow soon.
```
|- controllers/
|-- app.go
|- views/
|-- app/
|--- index.html
| main.go
```

*controllers/app.go*
```
package controller

import (
	"github.com/tsurai/amber"
)

type App struct {
  *amber.Controller
}

func (c *App) Index() {
	c.Render()
}
```

*views/app/index.html*
```
<h1>Hello Controller</h1>
```
  
*main.go*
```
package main

import (
	"github.com/tsurai/amber"
	c "controllers"
)

func main() {
	a := amber.New()

	a.AddRoute("GET", "/", (*c.App).Index)

	a.Run()
}

```

## Todo List
- proper error handling
- models & db interfaces
- middleware
- convert post data to struct
- flexible configurations

