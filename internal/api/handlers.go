package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"WorldOfWarcraft/internal/domain"
)

type createCharacterRequest struct {
	Name  domain.Name  `json:"name"`
	Race  domain.Race  `json:"race"`
	Class domain.Class `json:"class"`
}

func (s *Server) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	var req createCharacterRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest,
			errorResponse{Error: "malformed JSON body: " + err.Error()})
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest,
			errorResponse{Error: "body must contain a single JSON object"})
		return
	}

	c, err := s.store.Create(req.Name, req.Race, req.Class)
	if err != nil {
		writeErr(w, err)
		return
	}

	w.Header().Set("Location", "/characters/"+url.PathEscape(string(c.Name)))
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) handleGetCharacter(w http.ResponseWriter, r *http.Request) {
	name, err := domain.NormalizeName(domain.Name(r.PathValue("name")))
	if err != nil {
		writeErr(w, err)
		return
	}

	c, err := s.store.Get(name)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleListCharacters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
