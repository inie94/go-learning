package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	api "task-manager/internal/api/generated"
	middleware "task-manager/internal/api/middleware"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create generated server.
	srv, err := api.NewServer(api.UnimplementedHandler{}, 
		api.WithMiddleware(middleware.Recovery(logger)),
		api.WithMiddleware(middleware.Logger(logger)),
	)

	if err != nil {
		log.Fatal(err)
	}
	if err := http.ListenAndServe(":8080", srv); err != nil {
		log.Fatal(err)
	}

}
