package server

import "net/http"

type Server struct {
	server *http.Server
	mux    *http.ServeMux
}

func NewServer() *Server {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("web")))

	return &Server{
		server: &http.Server{
			Addr:    ":3000",
			Handler: mux,
		},
		mux: mux,
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Mux() *http.ServeMux {
	return s.mux
}
