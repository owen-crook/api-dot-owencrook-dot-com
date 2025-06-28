package main

import (
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/config"
	"github.com/owen-crook/api-dot-owencrook-dot-com/internal/router"
)

func main() {
	cfg := config.LoadConfig()

	// Setup router
	r := router.SetupRouter(cfg)

	r.Run()
}
