package lb

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type (
	LB struct {
		port     int
		listener net.Listener
		wg       sync.WaitGroup
		shutdown chan struct{}
	}
)

func New(port int) *LB {
	return &LB{
		port:     port,
		shutdown: make(chan struct{}),
	}
}

func (lb *LB) Start() error {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", lb.port))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	lb.listener = listener

	for {
		select {
		case <-lb.shutdown:
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-lb.shutdown:
					return nil
				default:
					log.Printf("Failed to accept connection: %v", err)
					continue
				}
			}

			lb.wg.Add(1)
			go func() {
				defer lb.wg.Done()
				lb.HandleConnection(conn)
			}()
		}
	}
}

func (lb *LB) Shutdown() error {
	close(lb.shutdown)

	if lb.listener != nil {
		if err := lb.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	lb.wg.Wait()
	return nil
}

func (lb *LB) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Set connection timeout
	duration, _ := time.ParseDuration("10m")
	deadline := time.Now().Add(duration)
	if err := conn.SetDeadline(deadline); err != nil {
		log.Printf("failed to set deadline: %v", err)
		return
	}

	backendAddr := "127.0.0.1:8080"
	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("failed to connect to BE server: %v\n", err)
		return
	}
	defer backendConn.Close()

	go io.Copy(backendConn, conn)
	io.Copy(conn, backendConn)
}
