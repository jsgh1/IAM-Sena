package main

import (
	"log"
	"net/http"
	"time"

	"iamapp/backend/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: application.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("IAM API escuchando en :%s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
