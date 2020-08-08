package mud

import (
	"encoding/json"
	"fmt"
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
)

type contentType string

const (
	contentTypeMUD       contentType = "application/mud+json"
	contentTypeSignature             = "application/pkcs7-signature"
	contentTypeUnknown               = "unknown"
	contentTypeInvalid               = "invalid"
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
	ValidateHeaders *bool `json:"validate_headers,omitempty"`
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

	w.Header().Set("server", "MUD File Server") // TODO: make this optional / configurable? Add version?

	fmt.Println(r)
	fmt.Println(r.URL.Path)

	replacer := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	root := replacer.ReplaceAll(m.Root, ".")
	suffix := replacer.ReplaceAll(r.URL.Path, "")
	filename := sanitizedPathJoin(root, suffix)

	fmt.Println(filename)

	// get information about the file
	info, err := os.Stat(filename)
	if err != nil {
		// TODO: perform better checks? not exists vs permission error?
		return m.notFound(w, r, next)
	}

	fmt.Println(info)

	if info.IsDir() {
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	contentType, err := m.detectContentType(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	fmt.Println(contentType)

	switch contentType {
	case contentTypeSignature:
		fmt.Println("send signature")
	case contentTypeMUD:
		if !m.validHeaders(r) {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
		fmt.Println("send MUD")
		// TODO: validate file is valid MUD (configurable?)
		// TODO: validate file has valid signature (configurable? only when it's available in this server too?)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("can't send unknown/invalid content type")
	}

	file, err := m.openFile(filename)
	if err != nil {
		return m.notFound(w, r, next)
	}
	defer file.Close()

	w.Header().Set("ETag", m.calculateETag(info))
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", string(contentType))
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)

	return nil
}

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
// and joins it to the (server) root
//
// source: Caddy
func sanitizedPathJoin(root, reqPath string) string {

	if root == "" {
		root = "."
	}

	return filepath.Join(root, filepath.FromSlash(path.Clean("/"+reqPath)))
}

// notFound returns a 404 error or, if pass-thru is enabled,
// it calls the next handler in the chain.
//
// source: Caddy
func (m *FileServer) notFound(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
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
		// if the file can be unmarshalled as JSON, this may be a MUD file
		return contentTypeMUD, nil
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

// calculateEtag creates an ETag from file metadata in os.FileInfo
//
// source: Caddy
func (m *FileServer) calculateETag(d os.FileInfo) string {
	t := strconv.FormatInt(d.ModTime().Unix(), 36)
	s := strconv.FormatInt(d.Size(), 36)
	return `"` + t + s + `"`
}
