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

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS
		w.Header().Set("Access-Control-Allow-Origin", "http://mia-preview-1.s3-website.eu-west-2.amazonaws.com") // Frontend URL
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")                              // Allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")                            // Include Authorization header

		if r.Method == http.MethodOptions {
			// Respond to preflight requests
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	s.router.HandleFunc("POST /business-summaries", handlers.BusinessSummaries(s.config.Store, s.config.Researcher, s.config.ImagesClient))
	s.router.HandleFunc("PATCH /business-summaries", handlers.PatchBusinessSummaries(s.config.Store))
	s.router.HandleFunc("GET /business-summaries", handlers.GetBusinessSummaries(s.config.Store))

	s.router.HandleFunc("GET /sitemap", handlers.GetSitemap(s.config.Store))

	s.router.HandleFunc("POST /campaigns", handlers.GenerateCampaigns(s.config.Store, s.config.CampaignClient))
	s.router.HandleFunc("GET /campaigns/{id}", handlers.GetCampaign(s.config.Store))
}

func (s *Server) Start() error {
	stack := CreateMiddlewareStack(
		s.corsMiddleware, // CORS middleware should be first
		Auth,
		Logging,
	)

	return http.ListenAndServe(s.listenAddr, stack(s.router))
}
