package common

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

func NewRoute(method, pattern string, handler http.HandlerFunc) Route {
	return Route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

type Route struct {
	Method  string
	Regex   *regexp.Regexp
	Handler http.HandlerFunc
}

func HandleRoute(routes []Route, w http.ResponseWriter, r *http.Request) bool {

	var allow []string
	for _, route := range routes {
		matches := route.Regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.Method {
				allow = append(allow, route.Method)
				continue
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, matches[1:])
			route.Handler(w, r.WithContext(ctx))
			return true
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return true
	}

	return false
}

type ctxKey struct{}

func GetField(r *http.Request, index int) string {
	fields := r.Context().Value(ctxKey{}).([]string)
	return fields[index]
}
