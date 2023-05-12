package actuator

import (
	"net/http"
	"runtime/pprof"
)

type DebugHandler struct {
}

func NewDebugHandler() *DebugHandler {
	return &DebugHandler{}
}

func (v *DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/text")

	// stack traces of all current goroutines
	w.Write([]byte("------------------------------------------------\n"))
	pprof.Lookup("goroutine").WriteTo(w, 2)

	// a sampling of all heap allocations
	w.Write([]byte("------------------------------------------------\n"))
	pprof.Lookup("heap").WriteTo(w, 1)

	// stack traces that led to blocking on synchronization primitives
	w.Write([]byte("------------------------------------------------\n"))
	pprof.Lookup("block").WriteTo(w, 1)
}
