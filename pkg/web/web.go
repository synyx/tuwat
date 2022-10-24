package web

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	html "html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/synyx/gonagdash/pkg/aggregation"
	"github.com/synyx/gonagdash/pkg/buildinfo"
	"github.com/synyx/gonagdash/pkg/config"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

//go:embed templates/*
var templates embed.FS

type webHandler struct {
	routes []route
	fs     fs.FS

	aggregator  *aggregation.Aggregator
	environment string
}

type webContent struct {
	Version     string
	Environment string
	Content     any
}

func WebHandler(cfg *config.Config, aggregator *aggregation.Aggregator) http.Handler {
	handler := &webHandler{
		aggregator:  aggregator,
		environment: cfg.Environment,
	}

	if cfg.Mode == "dev" {
		_, filename, _, _ := runtime.Caller(0)
		templatespath := path.Join(path.Dir(filename), "/templates")
		handler.fs = os.DirFS(templatespath)
	} else {
		handler.fs, _ = fs.Sub(templates, "templates")
	}

	handler.routes = []route{
		newRoute("GET", "/", handler.alerts),
		newRoute("GET", "/alerts", handler.alerts),
		newRoute("POST", "/alerts/([^/]+)/silence", handler.silence),
		newRoute("GET", "/ws/alerts", websocket.Handler(handler.wsalerts).ServeHTTP),
		newRoute("GET", "/sse/alerts", handler.ssealerts),
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

	templateFiles = append([]string{"_base.gohtml"}, patterns...)
	templateDefinition = "base"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		otelzap.Ctx(req.Context()).Error("compiling template failed", zap.Error(err))
		panic(err)
	}

	return func(w http.ResponseWriter, statusCode int, data webContent) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(statusCode)

		data.Version = buildinfo.Version
		data.Environment = h.environment

		err := tmpl.ExecuteTemplate(w, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(req.Context()).Error("template execution failed", zap.Error(err))
			panic(err)
		}
	}
}

func (h *webHandler) partialRenderer(req *http.Request, patterns ...string) renderFunc {
	var templateFiles []string
	var templateDefinition string

	templateFiles = append([]string{"_stream.gohtml"}, patterns...)
	templateDefinition = "base"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		otelzap.Ctx(req.Context()).Error("compiling template failed", zap.Error(err))
		panic(err)
	}

	return func(w http.ResponseWriter, statusCode int, data webContent) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
		w.WriteHeader(statusCode)

		data.Version = buildinfo.Version
		data.Environment = h.environment

		err := tmpl.ExecuteTemplate(w, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(req.Context()).Error("template execution failed", zap.Error(err))
			panic(err)
		}
	}
}

type sseRenderFunc func(data webContent)

func (h *webHandler) sseRenderer(w http.ResponseWriter, req *http.Request, patterns ...string) sseRenderFunc {
	var templateFiles []string
	var templateDefinition string

	templateFiles = append([]string{"_stream.gohtml"}, patterns...)
	templateDefinition = "content-container"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		otelzap.Ctx(req.Context()).Error("compiling template failed", zap.Error(err))
		panic(err)
	}

	// prepare the flusher
	flusher, _ := w.(http.Flusher)
	ctx, _ := context.WithTimeout(req.Context(), 10*time.Minute)
	req = req.WithContext(ctx)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	fmt.Fprint(w, "retry: 60000\n\n")

	flusher.Flush()

	return func(data webContent) {
		data.Version = buildinfo.Version
		data.Environment = h.environment

		buf := new(bytes.Buffer)

		tr := trace.SpanFromContext(req.Context())
		fmt.Fprintf(w, "id: %s\n", tr.SpanContext().TraceID())

		fmt.Fprint(w, "event: message\n")
		err = tmpl.ExecuteTemplate(buf, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(req.Context()).Info("template execution failed", zap.Error(err))
			panic(err)
		}

		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			_, err = w.Write([]byte("data: "))
			_, err = w.Write(scanner.Bytes())
			_, err = w.Write([]byte("\n"))
			if err != nil {
				otelzap.Ctx(req.Context()).Info("template execution failed", zap.Error(err))
				panic(err)
			}
		}
		_, err = w.Write([]byte("\n"))
		flusher.Flush()
	}
}

type wsRenderFunc func(data webContent)

func (h *webHandler) wsRenderer(s *websocket.Conn, patterns ...string) wsRenderFunc {
	var templateFiles []string
	var templateDefinition string

	templateFiles = append([]string{"_stream.gohtml"}, patterns...)
	templateDefinition = "content-container"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		otelzap.Ctx(s.Request().Context()).Error("compiling template failed", zap.Error(err))
		panic(err)
	}

	return func(data webContent) {
		w, err := s.NewFrameWriter(websocket.TextFrame)
		if err != nil {
			panic(err)
		}

		data.Version = buildinfo.Version
		data.Environment = h.environment

		buf := new(bytes.Buffer)

		err = tmpl.ExecuteTemplate(buf, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(s.Request().Context()).Info("template execution failed", zap.Error(err))
			panic(err)
		}

		_, err = w.Write(buf.Bytes())
		if err != nil {
			otelzap.Ctx(s.Request().Context()).Info("sending failed", zap.Error(err))
			panic(err)
		}
	}
}

type ctxKey struct{}

func getField(r *http.Request, index int) string {
	fields := r.Context().Value(ctxKey{}).([]string)
	return fields[index]
}

func niceDuration(d time.Duration) string {
	if d > 2*time.Hour*24 {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	} else if d > 2*time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	} else if d > 2*time.Minute {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d > 0 {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else {
		return d.String()
	}
}

func formatJson(s any) string {
	x, _ := json.MarshalIndent(s, "", "  ")
	return string(x)
}
