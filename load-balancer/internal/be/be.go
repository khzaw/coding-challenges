package be

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type (
	BE struct {
		port   int
		server *http.Server
	}
)

func New(port int) *BE {
	return &BE{port: port}
}

func (be *BE) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received a request from %+v", r.Method)
		fmt.Fprintf(w, "Hello from server at :%d!", be.port)
	})

	be.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", be.port),
		Handler: nil,
	}

	return be.server.ListenAndServe()
}

func (be *BE) Shutdown(ctx context.Context) error {
	if be.server == nil {
		return nil
	}

	return be.server.Shutdown(ctx)
}
