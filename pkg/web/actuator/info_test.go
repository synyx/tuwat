package actuator

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInfoReq(t *testing.T) {
	vh := NewVersionHandler()

	req, err := http.NewRequest("GET", "/actuator/info", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(vh.ServeHTTP)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "application") {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), "application:tuwat")
	}
}
