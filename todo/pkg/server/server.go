package server

import (
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	server *http.Server
	mux    *http.ServeMux
}

func NewServer() *Server {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	path := webPath()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(path)))

	return &Server{
		server: &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		},
		mux: mux,
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func webPath() string {
	path := "web"

	if exePath, err := os.Executable(); err == nil {
		webLocalPath := filepath.Join((filepath.Dir(exePath)), path)

		if info, err := os.Stat(webLocalPath); err == nil && info.IsDir() {
			path = webLocalPath
			return path
		}
	}

	webPath, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	return webPath
}

func (s *Server) Mu() *http.ServeMux {
	return s.mux
}
