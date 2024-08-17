package api

import (
	"encoding/json"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/storage"
)

type Server struct {
	listenAddr string
	store      storage.Storage
	router     *http.ServeMux
}

func NewServer(listenAddr string, store storage.Storage) *Server {
	s := &Server{listenAddr: listenAddr, store: store, router: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("POST /users/{id}", s.handleGetUserByID)
}

func (s *Server) Start() error {
	stack := CreateMiddlewareStack(
		Logging,
	)

	return http.ListenAndServe(s.listenAddr, stack(s.router))
}

func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := s.store.GetUserByID(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}
