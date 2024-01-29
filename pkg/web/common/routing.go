package common

import (
	"net/http"
	"regexp"
)

func NewRoute(method, pattern string, handler http.HandlerFunc) Route {
	return Route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

type Route struct {
	Method  string
	Regex   *regexp.Regexp
	Handler http.HandlerFunc
}
