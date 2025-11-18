package web

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	html "html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/websocket"

	"github.com/synyx/tuwat/pkg/aggregation"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/version"
	"github.com/synyx/tuwat/pkg/web/common"
)

//go:embed templates/*
var templates embed.FS

type WebHandler struct {
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

var (
	TemplateError   = errors.New("template execution failed")
	DisconnectError = errors.New("client disconnected")
)

func NewWebHandler(cfg *config.Config, aggregator *aggregation.Aggregator) *WebHandler {
	handler := &WebHandler{
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

func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				if errors.Is(err, DisconnectError) {
					slog.DebugContext(r.Context(), "panic serving", slog.Any("error", err))
				} else {
					slog.InfoContext(r.Context(), "panic serving", slog.Any("error", err))
				}
			default:
				slog.ErrorContext(r.Context(), "panic serving", slog.Any("error", err))
			}

			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		}
	}()

	if ok := common.HandleRoute(h.routes, w, r); !ok {
		h.notFound(w, r)
	}
}

type renderFunc func(w http.ResponseWriter, statusCode int, data webContent)

func (h *WebHandler) baseRenderer(req *http.Request, dashboardName string, templateFiles ...string) renderFunc {
	templateDefinition := "base"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		slog.ErrorContext(req.Context(), "compiling template failed", slog.Any("error", err))
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

		if err := tmpl.ExecuteTemplate(w, templateDefinition, data); err != nil {
			panic(errors.Join(TemplateError, err))
		}
	}
}

func (h *WebHandler) partialRenderer(req *http.Request, dashboardName string, templateFiles ...string) renderFunc {
	templateDefinition := "base"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		slog.ErrorContext(req.Context(), "compiling template failed", slog.Any("error", err))
		panic(err)
	}

	return func(w http.ResponseWriter, statusCode int, data webContent) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
		w.WriteHeader(statusCode)

		data.Version = version.Info.Version
		data.Environment = h.environment
		data.Style = h.style
		data.Dashboard = dashboardName

		if err := tmpl.ExecuteTemplate(w, templateDefinition, data); err != nil {
			panic(errors.Join(TemplateError, err))
		}
	}
}

type sseRenderFunc func(data webContent) error

func (h *WebHandler) sseRenderer(w http.ResponseWriter, req *http.Request, patterns ...string) (sseRenderFunc, context.CancelFunc) {
	templateFiles := append([]string{"_stream.gohtml"}, patterns...)
	templateDefinition := "content-container"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		slog.ErrorContext(req.Context(), "compiling template failed", slog.Any("error", err))
		return func(data webContent) error {
			return err
		}, func() {}
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

	if _, err := fmt.Fprint(w, "retry: 60000\n\n"); err != nil {
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

		if err := tmpl.ExecuteTemplate(buf, templateDefinition, data); err != nil {
			panic(errors.Join(TemplateError, err))
		}

		tr := trace.SpanFromContext(req.Context())
		if _, err := fmt.Fprintf(w, "id: %s\n", tr.SpanContext().TraceID()); err != nil {
			panic(errors.Join(DisconnectError, err))
		}
		if _, err := fmt.Fprint(w, "event: message\n"); err != nil {
			panic(errors.Join(DisconnectError, err))
		}

		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			if _, err := w.Write([]byte("data: ")); err != nil {
				panic(errors.Join(DisconnectError, err))
			}
			if _, err := w.Write(scanner.Bytes()); err != nil {
				panic(errors.Join(DisconnectError, err))
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				panic(errors.Join(DisconnectError, err))
			}
		}

		if _, err := w.Write([]byte("\n")); err != nil {
			panic(errors.Join(DisconnectError, err))
		}
		flusher.Flush()

		return nil
	}, cancel
}

type wsRenderFunc func(data webContent)

func (h *WebHandler) wsRenderer(s *websocket.Conn, patterns ...string) wsRenderFunc {
	templateFiles := append([]string{"_stream.gohtml"}, patterns...)
	templateDefinition := "content-container"

	funcs := html.FuncMap{
		"niceDuration": niceDuration,
		"json":         formatJson,
	}
	tmpl := html.New(templateDefinition).Funcs(funcs)
	tmpl, err := tmpl.ParseFS(h.fs, templateFiles...)
	if err != nil {
		panic(errors.Join(TemplateError, err))
	}

	return func(data webContent) {
		w, err := s.NewFrameWriter(websocket.TextFrame)
		if err != nil {
			panic(fmt.Errorf("cannot create websocket text frame writer: %w", err))
		}

		data.Version = version.Info.Version
		data.Environment = h.environment
		data.Style = h.style

		if err := tmpl.ExecuteTemplate(w, templateDefinition, data); err != nil {
			panic(errors.Join(TemplateError, err))
		}
	}
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
