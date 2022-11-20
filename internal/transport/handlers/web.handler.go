package transport

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/archellir/bookmark.arcbjorn.com/web"
)

type WebHandler struct{}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

func (handler *WebHandler) HandleStaticFiles(w http.ResponseWriter, r *http.Request) {
	distSubfolder, _ := fs.Sub(web.EmbededFilesystem, "dist")
	httpFileSystem := http.FileServer(http.FS(distSubfolder))
	httpFileSystem.ServeHTTP(w, r)
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
