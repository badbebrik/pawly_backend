package main

import (
	"chat/internal/app"
	"chat/internal/config"
	"log"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("create chat app: %v", err)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("run chat app: %v", err)
	}
}
