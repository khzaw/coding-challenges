package lb

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type (
	LB struct {
		port       int
		servers    []string
		currentIdx int
		listener   net.Listener
		wg         sync.WaitGroup
		shutdown   chan struct{}
	}
)

func New(port int, serverArgs []string) *LB {
	lb := &LB{
		port:       port,
		currentIdx: 0,
		shutdown:   make(chan struct{}),
	}

	var servers []string
	for _, a := range serverArgs {
		port, err := strconv.Atoi(a)
		if err != nil || port == lb.port || port < 0 || port > 65535 {
			log.Fatalf("invalid port: %v", err)
		}
		servers = append(servers, "127.0.0.1:"+a)
	}
	lb.servers = servers

	return lb
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
			go func(conn net.Conn) {
				defer lb.wg.Done()
				lb.HandleConnection(conn)
			}(conn)
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

	backendConn, err := net.Dial("tcp", lb.servers[lb.currentIdx])
	if err != nil {
		log.Printf("failed to connect to BE server: %v\n", err)
		return
	}
	defer backendConn.Close()
	lb.currentIdx = (lb.currentIdx + 1) % len(lb.servers)

	go io.Copy(backendConn, conn)
	io.Copy(conn, backendConn)
}
