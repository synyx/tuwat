package web

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

//go:embed static
var content embed.FS

func newNoListingFileServer() http.Handler {
	var staticFS fs.FS

	if dir, ok := os.LookupEnv("TUWAT_STATICDIR"); ok {
		if dir == "" {
			_, filename, _, _ := runtime.Caller(1)
			dir = path.Join(path.Dir(filename), "/static")
		}
		staticFS = os.DirFS(dir)
	} else {
		staticFS, _ = fs.Sub(content, "static")
	}

	return http.FileServer(noListingFS{http.FS(staticFS)})
}

type noListingFS struct {
	fs http.FileSystem
}

func (nfs noListingFS) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	if s, err := f.Stat(); err == nil && s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
