package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var port string = "9977"
var backupState sync.Map

func main() {
	router := NewRouter()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		// Catch Terminal (ctr-c) and SigTerm
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		// Signal recieved, shutdown
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Println("Service shutdown failed! %v", err)
		}

		log.Println("Stopping service on port [" + port + "]")
		close(idleConnsClosed)
	}()

	log.Println("Starting service on port [" + port + "]")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Println("Service start failed! %v", err)
	}

	<-idleConnsClosed
}
