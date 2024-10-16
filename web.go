package webServe

import (
	"crypto/md5"
	"errors"
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

type Config struct {
	Interceptor         gin.HandlerFunc
	DisableCacheControl bool
}

func NewHandler(fe fs.FS, conf Config) (gin.HandlerFunc, error) {
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

		if conf.Interceptor != nil {
			conf.Interceptor(c)
			if c.IsAborted() {
				return
			}
		}

		f, err := fe.Open(strings.TrimPrefix(c.Request.URL.Path, "/"))
		if err != nil {
			var fsError *fs.PathError
			if errors.As(err, &fsError) {
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
			}
		}
		_ = f.Close()

		if !conf.DisableCacheControl {
			c.Header("Cache-Control", "public, max-age=2592000, immutable")
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}, nil
}

func New(fe fs.FS) (gin.HandlerFunc, error) {
	return NewHandler(fe, Config{})
}
