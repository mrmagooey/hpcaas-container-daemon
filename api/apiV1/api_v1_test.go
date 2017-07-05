package apiV1

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCodeConfigurationV1_POST(t *testing.T) {
	var jsonStr = []byte(`{"codeParameters":[{"key":"foo", "value":"bar"}, {"key":"hello", "value": "world"}]}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SetCodeConfig())
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"status": "success", "data":{"message":"parameter accepted"}}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
