package mud

import "github.com/caddyserver/caddy/v2"

func init() {
	caddy.RegisterModule(MudFileServer{})
}

type MudFileServer struct {
}

func (MudFileServer) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.mud",
		New: func() caddy.Module { return new(MudFileServer) },
	}
}
