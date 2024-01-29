package web

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/synyx/tuwat/pkg/web/common"
	html "html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/synyx/tuwat/pkg/aggregation"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/version"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

//go:embed templates/*
var templates embed.FS

type webHandler struct {
	routes []common.Route
	fs     fs.FS

	aggregator  *aggregation.Aggregator
	environment string
	style       string
	dashboards  map[string]*config.Dashboard
}

type webContent struct {
	Version     string
	Environment string
	Content     any
	Dashboards  map[string]*config.Dashboard
	Dashboard   string
	Style       string
}

func WebHandler(cfg *config.Config, aggregator *aggregation.Aggregator) http.Handler {
	handler := &webHandler{
		aggregator:  aggregator,
		environment: cfg.Environment,
		style:       cfg.Style,
		dashboards:  cfg.Dashboards,
	}

	if dir, ok := os.LookupEnv("TUWAT_TEMPLATEDIR"); ok {
		if dir == "" {
			_, filename, _, _ := runtime.Caller(0)
			dir = path.Join(path.Dir(filename), "/templates")
		}
		handler.fs = os.DirFS(dir)
	} else {
		handler.fs, _ = fs.Sub(templates, "templates")
	}

	handler.routes = []common.Route{
		common.NewRoute("GET", "/", handler.alerts),
		common.NewRoute("GET", "/foo.php", http.RedirectHandler("/alerts/foo.php", http.StatusSeeOther).ServeHTTP),
		common.NewRoute("GET", "/alerts/([^/]+)", handler.alerts),
		common.NewRoute("GET", "/ws/(?:alerts/([^/]+))?", websocket.Handler(handler.wsalerts).ServeHTTP),
		common.NewRoute("GET", "/sse/(?:alerts/([^/]+))?", handler.ssealerts),
		common.NewRoute("POST", "/alerts/([^/]+)/silence", handler.silence),
	}

	return handler
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
		matches := route.Regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.Method {
				allow = append(allow, route.Method)
				continue
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, matches[1:])
			route.Handler(w, r.WithContext(ctx))
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

func (h *webHandler) baseRenderer(req *http.Request, dashboardName string, patterns ...string) renderFunc {
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

		data.Version = version.Info.Version
		data.Environment = h.environment
		data.Style = h.style
		data.Dashboards = h.dashboards
		data.Dashboard = dashboardName

		err := tmpl.ExecuteTemplate(w, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(req.Context()).Error("template execution failed", zap.Error(err))
			panic(err)
		}
	}
}

func (h *webHandler) partialRenderer(req *http.Request, dashboardName string, patterns ...string) renderFunc {
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

		data.Version = version.Info.Version
		data.Environment = h.environment
		data.Style = h.style
		data.Dashboard = dashboardName

		err := tmpl.ExecuteTemplate(w, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(req.Context()).Error("template execution failed", zap.Error(err))
			panic(err)
		}
	}
}

type sseRenderFunc func(data webContent) error

func (h *webHandler) sseRenderer(w http.ResponseWriter, req *http.Request, patterns ...string) (sseRenderFunc, context.CancelFunc) {
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
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("sse: response writer does not implement flush interface")
	}
	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Minute)
	req = req.WithContext(ctx)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	_, err = fmt.Fprint(w, "retry: 60000\n\n")
	if err != nil {
		return func(data webContent) error {
			return err
		}, cancel
	}

	flusher.Flush()

	return func(data webContent) error {
		data.Version = version.Info.Version
		data.Environment = h.environment
		data.Style = h.style

		buf := new(bytes.Buffer)

		err = tmpl.ExecuteTemplate(buf, templateDefinition, data)
		if err != nil {
			otelzap.Ctx(req.Context()).Info("template execution failed", zap.Error(err))
			panic(err)
		}

		tr := trace.SpanFromContext(req.Context())
		_, err = fmt.Fprintf(w, "id: %s\n", tr.SpanContext().TraceID())
		_, err = fmt.Fprint(w, "event: message\n")

		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			_, err = w.Write([]byte("data: "))
			_, err = w.Write(scanner.Bytes())
			_, err = w.Write([]byte("\n"))

		}
		_, err = w.Write([]byte("\n"))
		flusher.Flush()

		if err != nil {
			otelzap.Ctx(req.Context()).Info("template execution failed", zap.Error(err))
		}
		return err
	}, cancel
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

		data.Version = version.Info.Version
		data.Environment = h.environment
		data.Style = h.style

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
