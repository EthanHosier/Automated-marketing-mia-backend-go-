package main

import (
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"github.com/ethanhosier/mia-backend-go/api"
	"github.com/ethanhosier/mia-backend-go/config"
	"github.com/joho/godotenv"
)

func main2() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	listenAddr := flag.String("listen", ":8080", "HTTP server listen address")
	flag.Parse()

	server := api.NewServer(*listenAddr, config.NewProdServerConfig())
	log.Printf("Starting server on %s", *listenAddr)
	log.Fatal(server.Start())
}

func main() {
	fmt.Println("Hello, World!")
}
