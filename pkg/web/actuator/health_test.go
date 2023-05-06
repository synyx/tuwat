package actuator

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthReq(t *testing.T) {
	vh := newHealthActuator()

	SetHealth("main", OutOfService, "test")
	SetHealth("main", Up, "test")

	req, err := http.NewRequest("GET", "/actuator/health", nil)
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

	if !strings.Contains(rr.Body.String(), string(vh.Status())) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), "status:UP")
	}
}
