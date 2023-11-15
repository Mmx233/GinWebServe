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

func NewWithInterceptor(fe fs.FS, handler gin.HandlerFunc) (gin.HandlerFunc, error) {
	file, err := fe.Open("index.html")
	if err != nil {
		return nil, err
	}
	fileContentBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	_ = file.Close()
	index := string(fileContentBytes)
	indexEtag := etag(fileContentBytes)

	fileServer := http.StripPrefix("/", http.FileServer(http.FS(fe)))

	return func(c *gin.Context) {
		if c.Request.Method != "GET" || c.Writer.Written() {
			return
		}

		if handler != nil {
			handler(c)
			if c.IsAborted() {
				return
			}
		}

		f, err := fe.Open(strings.TrimPrefix(c.Request.URL.Path, "/"))
		if err != nil {
			if _, ok := err.(*fs.PathError); ok {
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
				c.String(500, err.Error())
				return
			}
		}
		_ = f.Close()

		c.Header("Cache-Control", "public, max-age=2592000, immutable")
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}, nil
}

func New(fe fs.FS) (gin.HandlerFunc, error) {
	return NewWithInterceptor(fe, nil)
}
