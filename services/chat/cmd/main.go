package main

import (
	"chat/internal/app"
	"chat/internal/config"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load chat config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("create chat app: %v", err)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("run chat app: %v", err)
	}
}
