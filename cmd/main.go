package main

import (
	"github.com/pavel-trbv/go-todo-app/internal/handler"
	"github.com/pavel-trbv/go-todo-app/internal/server"
	"log"
)

func main() {
	handlers := new(handler.Handler)

	srv := new(server.Server)
	if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occured while running http server: %s", err)
	}
}
