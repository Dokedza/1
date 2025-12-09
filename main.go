package main

import (
	cfg "1/cfg"
	"1/checker"
	handlers "1/handler"
	"1/storage"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	conf := cfg.Load()

	store := storage.NewFileStorage(conf.StorePath)

	err := store.Restore()
	if err != nil {
		log.Println("Restore failed:", err)
	}
	linkChecker := checker.NewLinkChecker(store, conf.WorkersCount)
	linkChecker.Start()

	handler := handlers.NewHandler(store, linkChecker)
	rourer := handlers.NewRouter(handler)

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Port),
		Handler:      rourer,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting server on", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	linkChecker.Stop()

	err = store.Backup()
	if err != nil {
		log.Println("Backing up failed:", err)
	}

	err = server.Shutdown(ctx)
	if err != nil {
		log.Println("Server shutdown failed:", err)
	}
	log.Println("Server stopped")
}
