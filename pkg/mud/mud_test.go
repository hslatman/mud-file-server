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

func TestServeHTTP(t *testing.T) {

	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	boolValue := true
	s := &FileServer{
		Root:            path.Join(d, "./../../"), // we're executing in ./pkg/mud, so we need to move up directories to get to examples/
		ValidateHeaders: &boolValue,
		ValidateMUD:     &boolValue,
		SetETag:         &boolValue,
	}

	if s == nil {
		t.Fatal("MUD File Server not instantiated")
	}

	replacer := caddy.NewReplacer()
	newContext := context.WithValue(context.Background(), caddy.ReplacerCtxKey, replacer)
	req, err := http.NewRequestWithContext(newContext, "GET", "/examples/lightbulb2000.json", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/mud+json")
	req.Header.Set("Accept-Language", "*")
	req.Header.Set("User-Agent", "mud-file-server-test")

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
	if recorder.Header().Get("Server") != expectedServerHeader {
		t.Errorf("handler returned unexpected server header: got %v want %v",
			recorder.Header().Get("Server"), expectedServerHeader)
	}

	// TODO: some more checks?
	// TODO: some more tests?

}
