package web

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/synyx/gonagdash/pkg/buildinfo"
	"github.com/synyx/gonagdash/pkg/config"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

//go:embed templates
var templates embed.FS

type webHandler struct {
	routes []route
	fs     fs.FS
}

type webContent struct {
	Version string
	Content interface{}
}

func WebHandler(cfg *config.Config) http.Handler {
	handler := &webHandler{}

	if cfg.Mode == "dev" {
		_, filename, _, _ := runtime.Caller(1)
		templatespath := path.Join(path.Dir(filename), "/templates")
		handler.fs = os.DirFS(templatespath)
	} else {
		handler.fs, _ = fs.Sub(templates, "templates")
	}

	handler.routes = []route{
		newRoute("GET", "/", handler.alerts),
	}

	return handler
}

func newRoute(method, pattern string, handler http.HandlerFunc) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
}

func (h *webHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				otelzap.Ctx(r.Context()).Error("panic serving", zap.Error(err))
			default:
				otelzap.Ctx(r.Context()).Error("panic serving", zap.Any("error", err))
			}

			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)

		}
	}()

	var allow []string
	for _, route := range h.routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				allow = append(allow, route.method)
				continue
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, matches[1:])
			route.handler(w, r.WithContext(ctx))
			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.notFound(w, r)
}

type renderFunc func(w http.ResponseWriter, statusCode int, data webContent)

func (h *webHandler) baseRenderer(req *http.Request, patterns ...string) renderFunc {
	var templateFiles []string
	var templateDefinition string

	if req.Header.Get("Turbo-Frame") == "content-container" {
		templateFiles = append([]string{"_content.gohtml"}, patterns...)
		templateDefinition = "content-container"
	} else {
		templateFiles = append([]string{"_base.gohtml", "_content.gohtml"}, patterns...)
		templateDefinition = "base"
	}

	tmpl, err := template.ParseFS(h.fs, templateFiles...)
	if err != nil {
		otelzap.Ctx(req.Context()).Error("compiling template failed", zap.Error(err))
		panic(err)
	}

	return func(w http.ResponseWriter, statusCode int, data webContent) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(statusCode)

		data.Version = buildinfo.Version

		err := tmpl.ExecuteTemplate(w, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(req.Context()).Error("template execution failed", zap.Error(err))
			panic(err)
		}
	}
}

type ctxKey struct{}

func getField(r *http.Request, index int) string {
	fields := r.Context().Value(ctxKey{}).([]string)
	return fields[index]
}
