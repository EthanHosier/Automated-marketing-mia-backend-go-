package main

import (
	"flag"
	"log"

	"github.com/ethanhosier/mia-backend-go/api"
	"github.com/ethanhosier/mia-backend-go/storage"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "HTTP server listen address")
	flag.Parse()

	store := storage.NewMemoryStorage()

	server := api.NewServer(*listenAddr, store)
	log.Printf("Starting server on %s", *listenAddr)
	log.Fatal(server.Start())
}
