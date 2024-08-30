package main

import (
	"flag"
	"log"
	"os"

	"time"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
	// "github.com/sashabaranov/go-openai"

	"github.com/ethanhosier/mia-backend-go/api"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	listenAddr := flag.String("listen", ":8080", "HTTP server listen address")
	flag.Parse()

	store := storage.NewSupabaseStorage(NewSupabaseClient())
	llmClient := utils.CreateLLMClient()

	log.Println(listenAddr, store, llmClient)

	server := api.NewServer(*listenAddr, store, llmClient)

	go func() {
		utils.StartCanvaTokenRefresher(time.Minute * 30)
	}()

	log.Printf("Starting server on %s", *listenAddr)
	log.Fatal(server.Start())

	// resp, err := llmClient.OpenaiImageCompletion("Give me the index of the image which best represents a cat. This should just be a number, with starting index being 0, and nothing else", []string{"https://images.pexels.com/photos/45201/kitty-cat-kitten-pet-45201.jpeg", "https://cdn.britannica.com/79/232779-050-6B0411D7/German-Shepherd-dog-Alsatian.jpg"}, openai.GPT4o)

	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// log.Println(resp)
}

func NewSupabaseClient() *supa.Client {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	return supa.CreateClient(supabaseUrl, supabaseServiceKey)
}
