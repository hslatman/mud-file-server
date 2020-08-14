package mud

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/caddyserver/caddy/v2"
)

func setupFileServer() (*FileServer, error) {
	d, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	boolValue := true
	s := &FileServer{
		Root:            path.Join(d, "./../../"), // we're executing in ./pkg/mud, so we need to move up 2 directories to get to examples/
		ValidateHeaders: &boolValue,
		ValidateMUD:     &boolValue,
		SetETag:         &boolValue,
	}

	if s == nil {
		return nil, err
	}

	return s, nil
}

func prepareRequest(method, path string) (*http.Request, error) {
	replacer := caddy.NewReplacer()
	newContext := context.WithValue(context.Background(), caddy.ReplacerCtxKey, replacer)
	req, err := http.NewRequestWithContext(newContext, method, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/mud+json")
	req.Header.Set("Accept-Language", "*")
	req.Header.Set("User-Agent", "mud-file-server-test")

	return req, nil
}

func TestServeHTTP(t *testing.T) {

	s, err := setupFileServer()
	if err != nil {
		t.Fatal(err)
	}

	req, err := prepareRequest("GET", "/examples/lightbulb2000.json")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	err = s.ServeHTTP(recorder, req, nil)
	if err != nil {
		t.Error(err)
	}

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedServerHeader := `MUD File Server v0.1.0 (github.com/hslatman/mud-file-server)`
	if serverHeader := recorder.Header().Get("Server"); serverHeader != expectedServerHeader {
		t.Errorf("handler returned unexpected server header: got %v want %v",
			recorder.Header().Get("Server"), expectedServerHeader)
	}

	expectedContentTypeHeader := "application/mud+json"
	if contentTypeHeader := recorder.Header().Get("Content-Type"); contentTypeHeader != expectedContentTypeHeader {
		t.Errorf("handler returned unexpected content-type header: got %v want %v",
			recorder.Header().Get("Content-Type"), expectedContentTypeHeader)
	}

	// TODO: more checks?

}

func TestInvalidMUD(t *testing.T) {

	s, err := setupFileServer()
	if err != nil {
		t.Fatal(err)
	}

	req, err := prepareRequest("GET", "/examples/invalid.json")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	err = s.ServeHTTP(recorder, req, nil)
	if err != nil {
		t.Error(err)
	}

	expectedStatus := http.StatusInternalServerError
	if status := recorder.Code; status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expectedStatus)
	}

}

// TODO: some more tests?
