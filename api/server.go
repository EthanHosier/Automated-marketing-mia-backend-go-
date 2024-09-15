package api

import (
	"net/http"

	"github.com/ethanhosier/mia-backend-go/api/handlers"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
)

type Server struct {
	listenAddr string
	store      storage.Storage
	router     *http.ServeMux

	researcher *researcher.ResearcherClient
}

func NewServer(listenAddr string, store storage.Storage, researcher *researcher.ResearcherClient) *Server {
	s := &Server{
		listenAddr: listenAddr,
		store:      store,
		router:     http.NewServeMux(),
		researcher: researcher,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("POST /business-summaries", handlers.BusinessSummaries(s.store, s.researcher))
	s.router.HandleFunc("PATCH /business-summaries", handlers.PatchBusinessSummaries(s.store))
	s.router.HandleFunc("GET /business-summaries", handlers.GetBusinessSummaries(s.store))

	s.router.HandleFunc("GET /sitemap", handlers.GetSitemap(s.store))

	// s.router.HandleFunc("POST /campaigns", handlers.GenerateCampaigns(s.store, s.llmClient))
}

func (s *Server) Start() error {
	stack := CreateMiddlewareStack(
		Auth,
		Logging,
	)

	return http.ListenAndServe(s.listenAddr, stack(s.router))
}
