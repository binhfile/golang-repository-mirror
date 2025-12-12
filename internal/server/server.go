package server

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/example/go-mod-clone/internal/log"
)

type Server struct {
	storageRoot string
	host        string
	port        int
}

func NewServer(storageRoot, host string, port int) *Server {
	return &Server{
		storageRoot: storageRoot,
		host:        host,
		port:        port,
	}
}

func (s *Server) Start() error {
	log.Info("Starting Go module proxy server")
	log.Info("Storage root: %s", s.storageRoot)
	log.Info("Listening on %s:%d", s.host, s.port)

	// Create file server
	fs := http.FileServer(http.Dir(s.storageRoot))

	// Create router with custom handler
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleRequest(w, r, fs)
	}))

	// Start server
	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Info("Server started. Use GOPROXY=http://%s:%d go get ...", s.host, s.port)
	return server.ListenAndServe()
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request, fs http.Handler) {
	path := r.URL.Path

	// Log request
	log.Debug("Request: %s %s", r.Method, path)

	// Handle root path
	if path == "/" || path == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html>
<head><title>Go Module Proxy</title></head>
<body>
<h1>Go Module Proxy Server</h1>
<p>This is a Go module proxy server running from %s</p>
<p>Configure your Go environment:</p>
<pre>export GOPROXY=http://%s:%d
go get github.com/user/module@version</pre>
</body>
</html>`, filepath.Base(s.storageRoot), s.host, s.port)
		return
	}

	// Serve files from storage root
	fs.ServeHTTP(w, r)
}
