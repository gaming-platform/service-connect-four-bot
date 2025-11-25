package main

import (
	"fmt"
	"log"

	"github.com/gaming-platform/connect-four-bot/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Config:", cfg)
}
