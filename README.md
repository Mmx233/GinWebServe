# GinWebServe

Serve embeded SPA static website in gin

## Usage

```go
// /web/embed.go
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var dist embed.FS

func Fs() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}
```

```go
// /internal/router/frontend.go
package router

import (
	webServe "github.com/Mmx233/GinWebServe"
	"github.com/gin-gonic/gin"
	"your-project/web"
	"log"
)

func frontendHandler() gin.HandlerFunc {
	fs, err := web.Fs()
	if err != nil {
		log.Fatalln(err)
	}

	handler, err := webServe.New(fs)
	if err != nil {
		log.Fatalln(err)
	}

	return handler
}
```

```go
// /internal/router/init.go
package router

import (
	"github.com/gin-gonic/gin"
)

var E *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	E = gin.Default()

	E.Use(frontendHandler())
}

```
