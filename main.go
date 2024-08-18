package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"

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

	supabaseClient := NewSupabaseClient()
	store := storage.NewSupabaseStorage(supabaseClient)

	server := api.NewServer(*listenAddr, store)
	// log.Printf("Starting server on %s", *listenAddr)
	// log.Fatal(server.Start())
	fmt.Println(server)
	fmt.Println(utils.RunBedrock("Hello, world!"))
}

func NewSupabaseClient() *supa.Client {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	return supa.CreateClient(supabaseUrl, supabaseServiceKey)
}
