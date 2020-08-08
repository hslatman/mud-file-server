package mud

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
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

	if !m.validHeaders(r) {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	// TODO: determine requested file
	// TODO: validate headers (configurable?)
	// TODO: validate file exists; is no directory; etc.
	// TODO: determine the type of file requested (MUD vs. its signature)
	// TODO: validate file has valid signature (configurable?)
	// TODO: validate file is valid MUD (configurable?)
	// TODO: respond with the file contents and set appropriate headers

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("HELLO from MUD File Server!"))

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
	}
	return true
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
