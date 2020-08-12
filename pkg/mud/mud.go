// Copyright 2020 Herman Slatman
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mud

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"go.mozilla.org/pkcs7"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"github.com/hslatman/mud.yang.go/pkg/mudyang"
)

type contentType string

const (
	contentTypeMUD       contentType = "application/mud+json"
	contentTypeSignature contentType = "application/pkcs7-signature"
	contentTypeJSON      contentType = "application/json"
	contentTypeUnknown   contentType = "unknown"
	contentTypeInvalid   contentType = "application/octet-stream"
)

const (
	version string = "v0.1.0"
)

func init() {
	caddy.RegisterModule(FileServer{})
}

// FileServer implements a MUD File Server responder for Caddy.
type FileServer struct {
	// The path to the root directory with MUD files.
	// Default is `{http.vars.root}` if set; current working directory otherwise.
	Root string `json:"root,omitempty"`
	// Validate request headers according to https://www.rfc-editor.org/rfc/rfc8520
	// Default is true
	ValidateHeaders *bool `json:"validate_headers,omitempty"`
	// Validate the requested MUD file (if it exists)
	// Default is true
	ValidateMUD *bool `json:"validate_mud,omitempty"`
	// Set ETag header in responses
	// Defaults is true
	SetETag *bool `json:"set_etag,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (FileServer) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.mud_file_server",
		New: func() caddy.Module { return new(FileServer) },
	}
}

// Provision sets up the MUD File Server responder.
func (m *FileServer) Provision(ctx caddy.Context) error {

	if m.Root == "" {
		m.Root = "{http.vars.root}"
	}

	return nil
}

// ServeHTTP is the core handler for the MUD File Server.
func (m *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

	w.Header().Set("Server", "MUD File Server "+version+" (github.com/hslatman/mud-file-server)")

	replacer := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	root := replacer.ReplaceAll(m.Root, ".")
	suffix := replacer.ReplaceAll(r.URL.Path, "")
	filename := sanitizedPathJoin(root, suffix)

	// get information about the file
	info, err := os.Stat(filename)
	if err != nil {
		// TODO: perform better checks? not exists vs permission error?
		return m.notFound(w, r)
	}

	if info.IsDir() {
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	contentType, err := m.detectContentType(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	if contentType == contentTypeInvalid || contentType == contentTypeUnknown {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	if contentType == contentTypeJSON {
		if !m.validHeaders(r) {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
		var ok bool
		contentType, ok = m.validMUD(filename)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	file, err := m.openFile(filename)
	if err != nil {
		return m.notFound(w, r)
	}
	defer file.Close()

	m.setETag(w, info)
	m.setContentType(w, contentType)

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)

	return nil
}

// validHeaders validates the request headers
func (m *FileServer) validHeaders(r *http.Request) bool {
	if m.ValidateHeaders == nil || *m.ValidateHeaders {
		headers := r.Header

		accept, ok := headers["Accept"]
		if !ok {
			return false
		}
		if !contains(accept, "application/mud+json") {
			return false
		}

		if _, ok := headers["Accept-Language"]; !ok {
			return false
		}

		if _, ok := headers["User-Agent"]; !ok {
			return false
		}

		// TODO: add checks for empty Accept-Language and User-Agent?
	}
	return true
}

// validMUD validates the JSON to be a valid MUD according to RFC 8520
func (m *FileServer) validMUD(path string) (contentType, bool) {
	if m.ValidateMUD == nil || *m.ValidateMUD {
		json, err := ioutil.ReadFile(path)
		if err != nil {
			return contentTypeInvalid, false
		}
		mud := &mudyang.Mudfile{}
		if err := mudyang.Unmarshal([]byte(json), mud); err != nil {
			return contentTypeJSON, false
		}
		return contentTypeMUD, true
	}
	return contentTypeJSON, true
}

// setContentType sets the detected content type
func (m *FileServer) setContentType(w http.ResponseWriter, ct contentType) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", string(ct))
	}
}

// setETag sets an ETag based on an os.FileInfo object (if enabled)
func (m *FileServer) setETag(w http.ResponseWriter, info os.FileInfo) {
	if m.SetETag == nil || *m.SetETag {
		// implementation taken from github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver/staticfiles.go
		t := strconv.FormatInt(info.ModTime().Unix(), 36)
		s := strconv.FormatInt(info.Size(), 36)
		etag := `"` + t + s + `"`
		w.Header().Set("ETag", etag)
	}
}

// contains looks for a needle string in a haystack of strings
func contains(haystack []string, needle string) bool {
	needle = strings.ToLower(needle)
	for _, a := range haystack {
		if strings.ToLower(a) == needle {
			return true
		}
	}
	return false
}

// sanitizedPathJoin sanitizes the requested file path
// and joins it to the (server) root directory
//
// inspired by source: github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver/staticfiles.go
func sanitizedPathJoin(root, reqPath string) string {
	if root == "" {
		root = "."
	}
	return filepath.Join(root, filepath.FromSlash(path.Clean("/"+reqPath)))
}

// notFound returns a 404 error
//
// inspired by source: github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver/staticfiles.go
func (m *FileServer) notFound(w http.ResponseWriter, r *http.Request) error {
	return caddyhttp.Error(http.StatusNotFound, nil)
}

// detectContentType determines whether the requested file is a (potential) MUD file,
// a (potential) MUD signature or something different.
func (m *FileServer) detectContentType(path string) (contentType, error) {

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		// Any errors related to reading the file will be reported as an unknown filetype and err
		return contentTypeUnknown, err
	}

	if _, err := pkcs7.Parse(contents); err == nil {
		// if pkcs7 can parse without error, we assume this is a MUD signature file
		return contentTypeSignature, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(contents, &data); err == nil {
		// if the file can be unmarshalled as JSON, this _MAY_ be a MUD file
		return contentTypeJSON, nil
	}

	return contentTypeInvalid, nil
}

// openFile opens the file at the given filename.
func (m *FileServer) openFile(filename string) (*os.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}
