package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var staticFiles embed.FS

func Handler() http.Handler {
	files, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic(err)
	}
	return newHandler(files)
}

func newHandler(files fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(files))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(files, name); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		if path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}

		fallback := r.Clone(r.Context())
		fallbackURL := *r.URL
		fallbackURL.Path = "/"
		fallback.URL = &fallbackURL
		fileServer.ServeHTTP(w, fallback)
	})
}
