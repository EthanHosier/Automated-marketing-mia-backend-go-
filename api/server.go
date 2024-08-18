package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/types"
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

	// screenshotBase64, err := utils.PageScreenshot(utils.ScreenshotUrl + "?url=" + req.Url)

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// sitemap, err := utils.Sitemap(req.Url, 15)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	summaries, err := utils.BusinessPageSummaries(req.Url, 15, s.llmClient)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// resp := types.BusinessSummariesResponse{
	// 	Summaries: summaries,
	// }

	jsonData, err := json.Marshal(summaries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonString := string(jsonData)

	businessSummaries, err := utils.BusinessSummaryPoints(jsonString, s.llmClient)

	if err != nil {
		log.Println("Error getting business summary points:", err, ". Trying again (1st retry)")
		businessSummaries, err = utils.BusinessSummaryPoints(jsonString, s.llmClient)

		if err != nil {
			log.Println("Error getting business summary points:", err, ". Trying again (2nd retry)")
			businessSummaries, err = utils.BusinessSummaryPoints(jsonString, s.llmClient)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	resp := types.BusinessSummariesResponse{
		BusinessSummaries: *businessSummaries,
	}

	json.NewEncoder(w).Encode(resp)
}
