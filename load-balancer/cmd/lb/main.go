package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/khzaw/coding-challenges/load-balancer/internal/lb"
)

func main() {

	lb := lb.New(7070, os.Args[1:])

	errChan := make(chan error, 1)
	go func() {
		errChan <- lb.Start()
	}()

	// handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Load balancer error: %v", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		if err := lb.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}

	close(errChan)
	close(sigChan)
}
