package api

import (
	"net/http"

	"github.com/ethanhosier/mia-backend-go/api/handlers"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

type Server struct {
	listenAddr string
	store      storage.Storage
	router     *http.ServeMux
	llmClient  *utils.LLMClient
}

func NewServer(listenAddr string, store storage.Storage, llmClient *utils.LLMClient) *Server {
	s := &Server{listenAddr: listenAddr, store: store, router: http.NewServeMux(), llmClient: llmClient}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("POST /business-summaries", handlers.BusinessSummaries2(s.store, s.llmClient))
	s.router.HandleFunc("GET /business-summaries", handlers.GetBusinessSummaries(s.store))
	s.router.HandleFunc("GET /sitemap", handlers.GetSitemap(s.store))
	s.router.HandleFunc("POST /campaigns", handlers.GenerateCampaigns(s.store, s.llmClient))

}

func (s *Server) Start() error {
	stack := CreateMiddlewareStack(
		Auth,
		Logging,
	)

	return http.ListenAndServe(s.listenAddr, stack(s.router))
}
