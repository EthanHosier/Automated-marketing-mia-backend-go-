package main

import (
	"flag"
	"log"

	"github.com/ethanhosier/mia-backend-go/api"
	"github.com/ethanhosier/mia-backend-go/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	listenAddr := flag.String("listen", ":8080", "HTTP server listen address")
	flag.Parse()

	server := api.NewServer(*listenAddr, config.NewProdServerConfig())
	log.Printf("Starting server on %s", *listenAddr)
	log.Fatal(server.Start())
}

/*

Order of testing
1. The helper utils    [X]
2. The storage         [X]
3. Openai              [X] (other than openai client: TODO: abstract into interface)
2. The services 			 [X] ^ for services also
3. Canva 						   [X]
4. Researcher 				 [X]
5. Campaigns           []


NEED TO DO MOCK HTTP
*/
