package actuator

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProfiles(t *testing.T) {

	dh := NewDebugHandler()

	req, err := http.NewRequest("GET", "/actuator/debug", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(dh.ServeHTTP)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
