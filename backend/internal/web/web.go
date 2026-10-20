// Package web embeds the built frontend (Nuxt generate output) into the binary.
//
// The `dist` directory is populated by `make build` (copies frontend/.output/public
// here). During development it stays empty and only contains a .gitkeep; go:embed
// still compiles and the fallback handler simply returns 404 for static assets.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns an SPA-aware handler serving the embedded frontend.
// Unknown paths without a file extension fall back to index.html so
// client-side routing works.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("web: embed dist not available: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p != "" {
			if _, err := fs.Stat(sub, p); err != nil {
				// Not a real file: serve index.html for SPA routes.
				r2 := new(http.Request)
				*r2 = *r
				r2.URL.Path = "/"
				fileServer.ServeHTTP(w, r2)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
