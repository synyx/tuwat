package actuator

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

type oa3validatorRespWriter struct {
	buf        *bytes.Buffer
	statusCode int
	headers    http.Header
}

func (w *oa3validatorRespWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *oa3validatorRespWriter) Header() http.Header {
	return w.headers
}
func (w *oa3validatorRespWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func openapiValidationMiddleware(t *testing.T, next http.Handler, apiFile string) http.Handler {

	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(apiFile)
	if err != nil {
		t.Fatal(err)
	}

	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, err := router.FindRoute(r)
		if err != nil {
			t.Fatal(err)
		}

		// Validate request
		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
		}
		if err := openapi3filter.ValidateRequest(r.Context(), requestValidationInput); err != nil {
			t.Fatal(err)
		}

		// Wrap the response
		w2 := &oa3validatorRespWriter{buf: bytes.NewBufferString(""), headers: make(map[string][]string)}

		next.ServeHTTP(w2, r)

		// Transfer wrapped response up the chain
		for k, vv := range w2.headers {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(w2.statusCode)
		_, _ = w.Write(w2.buf.Bytes())

		// Validate response
		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 w2.statusCode,
			Header:                 w2.headers,
		}
		responseValidationInput.SetBodyBytes(w2.buf.Bytes())
		if err := openapi3filter.ValidateResponse(r.Context(), responseValidationInput); err != nil {
			t.Fatal(err)
		}
	})
}
