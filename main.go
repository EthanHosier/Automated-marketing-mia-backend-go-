package main

import "fmt"

// import (
// 	"flag"
// 	"github.com/joho/godotenv"
// 	supa "github.com/nedpals/supabase-go"
// 	"log"
// 	"os"
// 	"time"

// 	"github.com/ethanhosier/mia-backend-go/api"
// 	"github.com/ethanhosier/mia-backend-go/storage"
// 	"github.com/ethanhosier/mia-backend-go/utils"
// )

// func main() {
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatalf("Error loading .env file: %v", err)
// 	}

// 	listenAddr := flag.String("listen", ":8080", "HTTP server listen address")
// 	flag.Parse()

// 	store := storage.NewSupabaseStorage(NewSupabaseClient())
// 	llmClient := utils.CreateLLMClient()

// 	log.Println(listenAddr, store, llmClient)

// 	server := api.NewServer(*listenAddr, store, llmClient)

// 	go func() {
// 		utils.StartCanvaTokenRefresher(time.Minute * 30)
// 	}()

// 	log.Printf("Starting server on %s", *listenAddr)
// 	log.Fatal(server.Start())
// }

// func NewSupabaseClient() *supa.Client {
// 	supabaseUrl := os.Getenv("SUPABASE_URL")
// 	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_KEY")
// 	return supa.CreateClient(supabaseUrl, supabaseServiceKey)
// }

func main() {
	fmt.Println("yeahhh")
}
