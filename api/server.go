package api

import (
	"net/http"

	"github.com/ethanhosier/mia-backend-go/api/handlers"
	"github.com/ethanhosier/mia-backend-go/config"
)

type Server struct {
	listenAddr string
	config     config.ServerConfig
	router     *http.ServeMux
}

func NewServer(listenAddr string, config config.ServerConfig) *Server {
	s := &Server{
		listenAddr: listenAddr,
		config:     config,
		router:     http.NewServeMux(),
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("POST /business-summaries", handlers.BusinessSummaries(s.config.Store, s.config.Researcher))
	s.router.HandleFunc("PATCH /business-summaries", handlers.PatchBusinessSummaries(s.config.Store))
	s.router.HandleFunc("GET /business-summaries", handlers.GetBusinessSummaries(s.config.Store))

	s.router.HandleFunc("GET /sitemap", handlers.GetSitemap(s.config.Store))

	s.router.HandleFunc("POST /campaigns", handlers.GenerateCampaigns(s.config.Store, s.config.CampaignClient))
}

func (s *Server) Start() error {
	stack := CreateMiddlewareStack(
		Auth,
		Logging,
	)

	return http.ListenAndServe(s.listenAddr, stack(s.router))
}
