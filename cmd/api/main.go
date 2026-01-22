package main

import (
	"net/http"
	"time"

	"github.com/NurulloMahmud/medicalka-project/config"
	"github.com/NurulloMahmud/medicalka-project/internal/server"
)

func main() {
	cfg := config.Load()

	app, err := server.NewApplication(*cfg)
	if err != nil {
		panic(err)
	}
	defer app.DB.Close()

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.Logger.Printf("we are live on %s\n", "http://localhost"+cfg.ServerAddr)

	err = server.ListenAndServe()
	if err != nil {
		app.Logger.Fatal(err)
	}
}
