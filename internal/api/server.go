package api

import (
	"net/http"

	"WorldOfWarcraft/internal/store"
)

type Server struct {
	store store.CharacterStore
	mux   *http.ServeMux
}

func New(s store.CharacterStore) *Server {
	srv := &Server{store: s, mux: http.NewServeMux()}
	srv.routes()
	return srv
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /characters", s.handleCreateCharacter)
	s.mux.HandleFunc("GET /characters", s.handleListCharacters)
	s.mux.HandleFunc("GET /characters/{name}", s.handleGetCharacter)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
