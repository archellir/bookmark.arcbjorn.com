package transport

import (
	"fmt"
	"net/http"

	"github.com/archellir/bookmark.arcbjorn.com/web"
)

type WebHandler struct {
	staticFilesHandler http.Handler
}

func NewWebHandler(staticFilesHandler http.Handler) *WebHandler {
	return &WebHandler{
		staticFilesHandler: staticFilesHandler,
	}
}

func (handler *WebHandler) HandleStaticFiles(w http.ResponseWriter, r *http.Request) {
	handler.staticFilesHandler.ServeHTTP(w, r)
}

func (handler *WebHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// if r.URL.Path == "/favicon.ico" {
	// 	rawFile, _ := web.EmbededFilesystem.ReadFile("dist/favicon.ico")
	// 	w.Write(rawFile)
	// 	return
	// }

	rawFile, _ := web.EmbededFilesystem.ReadFile("dist/index.html")
	w.Write(rawFile)
}
