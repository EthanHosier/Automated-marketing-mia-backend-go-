package api

import (
	"encoding/json"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/ethanhosier/mia-backend-go/utils"
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
	// s.router.HandleFunc("POST /user", s.handleCreateUser)
	s.router.HandleFunc("POST /business-summaries", s.businessSummaries)
}

func (s *Server) Start() error {
	stack := CreateMiddlewareStack(
		Logging,
	)

	return http.ListenAndServe(s.listenAddr, stack(s.router))
}

// func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
// 	id := r.PathValue("id")
// 	user, err := s.store.GetUserByID(id)

// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(user)
// }

func (s *Server) businessSummaries(w http.ResponseWriter, r *http.Request) {
	var req types.BusinessSummariesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := types.ValidateBusinessSummariesRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	screenshotBase64, err := utils.GetPageScreenshot(utils.ScreenshotUrl + "?url=" + req.Url)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := types.BusinessSummariesResponse{
		ScreenshotBase64: screenshotBase64,
	}

	json.NewEncoder(w).Encode(resp)
}
