package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloReturnsServiceNameAndRandomString(t *testing.T) {
	handler := newHandler("backend-3")
	for _, tc := range []struct {
		method string
		path   string
		status int
		body   string
	}{
		{http.MethodGet, "/hello", http.StatusOK, "backend-3 " + RandomString + "\n"},
		{http.MethodGet, "/", http.StatusNotFound, "404 page not found\n"},
		{http.MethodPost, "/hello", http.StatusMethodNotAllowed, ""},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, nil))
		if recorder.Code != tc.status || recorder.Body.String() != tc.body {
			t.Errorf("%s %s: got status %d, body %q; want %d, %q", tc.method, tc.path, recorder.Code, recorder.Body.String(), tc.status, tc.body)
		}
	}
}
