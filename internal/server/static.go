package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:static
var staticFS embed.FS

func dashboardHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "dashboard not embedded", http.StatusInternalServerError)
		})
	}
	return http.FileServer(http.FS(sub))
}
