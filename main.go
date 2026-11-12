package main

import (
	"github.com/ZinollinIlyas/tarkov_oauth/internal/config"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/server"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("server run: %v", err)
	}
}
