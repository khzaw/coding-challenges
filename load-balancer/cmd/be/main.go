package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/khzaw/coding-challenges/load-balancer/internal/be"
)

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	port, err := strconv.Atoi(args[1])
	if err != nil || port < 0 || port > 65535 {
		fmt.Println("Invalid port. Using 8080")
		port = 8080
	}
	be := be.New(port)

	errChan := make(chan error, 1)
	go func() {
		errChan <- be.Start()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			log.Fatalf("BE error: %v", err)
			return err
		}
	case <-ctx.Done():

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
		defer shutdownCancel()

		if err := be.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown be: %w", err)
		}
	}

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
