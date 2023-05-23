package webServe

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

func etag(d []byte) string {
	hash := md5.New()
	hash.Write(d)
	return fmt.Sprintf("W/\"%x\"", hash.Sum(nil))
}

func New(fe fs.FS) (gin.HandlerFunc, error) {
	file, e := fe.Open("index.html")
	if e != nil {
		return nil, e
	}
	fileContentBytes, e := io.ReadAll(file)
	if e != nil {
		return nil, e
	}
	_ = file.Close()
	index := string(fileContentBytes)
	indexEtag := etag(fileContentBytes)

	fileServer := http.StripPrefix("/", http.FileServer(http.FS(fe)))

	return func(c *gin.Context) {
		f, e := fe.Open(strings.TrimPrefix(c.Request.URL.Path, "/"))
		if e != nil {
			if _, ok := e.(*fs.PathError); ok {
				if c.GetHeader("If-None-Match") == indexEtag {
					c.AbortWithStatus(304)
					return
				}
				c.Header("Content-Type", "text/html")
				c.Header("Cache-Control", "no-cache")
				c.Header("Etag", indexEtag)
				c.String(200, index)
				c.Abort()
				return
			} else {
				c.String(500, e.Error())
				return
			}
		}
		_ = f.Close()

		c.Header("Cache-Control", "public, max-age=2592000, immutable")
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}, nil
}
