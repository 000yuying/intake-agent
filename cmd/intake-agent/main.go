package main

import (
	"flag"
	"log"
	"github.com/yuying/intake-agent/internal/config"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("intake-agent starting on port %d", cfg.Server.Port)
}
