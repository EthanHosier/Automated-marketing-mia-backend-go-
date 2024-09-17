package main

import (
	"encoding/base64"
	"flag"
	"log"
	"os"

	"github.com/ethanhosier/mia-backend-go/api"
	"github.com/ethanhosier/mia-backend-go/config"
	"github.com/ethanhosier/mia-backend-go/images"
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
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	imgClient := config.NewProdServerConfig().ImagesClient
	result, err := imgClient.AiImageFrom("an annoying little sister in a green t shirt and mario stockings", images.StableImageCore)

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	err = os.WriteFile("image_result.png", result, 0644)
	if err != nil {
		log.Fatalf("Failed to write result to file: %v", err)
	}

	base64Result := base64.StdEncoding.EncodeToString(result)
	dataURL := "data:image/png;base64," + base64Result

	captions, err := imgClient.CaptionsFor(dataURL)

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Printf("Captions: %v", captions)
}
