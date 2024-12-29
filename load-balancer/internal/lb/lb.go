package lb

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type (
	server struct {
		addr    string
		healthy bool
	}
	LB struct {
		port       int
		mu         sync.RWMutex
		servers    []server
		currentIdx int
		listener   net.Listener
		wg         sync.WaitGroup
		shutdown   chan struct{}
	}
)

func New(port int, serverArgs []string) *LB {
	lb := &LB{
		port:       port,
		currentIdx: -1,
		shutdown:   make(chan struct{}),
	}

	var srvs []server
	for _, a := range serverArgs {
		port, err := strconv.Atoi(a)
		if err != nil || port == lb.port || port < 0 || port > 65535 {
			log.Fatalf("invalid port: %v", err)
		}
		srvs = append(srvs, server{addr: "127.0.0.1:" + a, healthy: true})
	}
	lb.servers = srvs

	return lb
}

func (lb *LB) Start() error {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", lb.port))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	lb.listener = listener

	ticker := time.NewTicker(5 * time.Second)
	go lb.runHealthChecks(ticker)

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

	nextServer, err := lb.getNextHealthyServer()
	if err != nil {
		log.Fatal("there is no healthy server")
	}

	backendConn, err := net.Dial("tcp", nextServer.addr)
	if err != nil {
		log.Printf("failed to connect to BE server: %v\n", err)
		return
	}
	defer backendConn.Close()

	go io.Copy(backendConn, conn)
	io.Copy(conn, backendConn)
}

func (lb *LB) getNextHealthyServer() (*server, error) {
	idx := (lb.currentIdx + 1) % len(lb.servers)
	for i := 0; i < len(lb.servers); i++ {
		idx := (idx + i) % len(lb.servers)
		if lb.servers[idx].healthy {
			lb.currentIdx = idx
			return &lb.servers[idx], nil
		}
	}

	return nil, errors.New("no healthy server")
}

func (lb *LB) runHealthChecks(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			for idx := range lb.servers {
				go lb.sendHealthCheck(idx)
			}
		case <-lb.shutdown:
			return
		}
	}
}

func (lb *LB) sendHealthCheck(idx int) error {
	res, err := http.Get("http://" + lb.servers[idx].addr)

	lb.mu.Lock()
	defer lb.mu.Unlock()
	if err != nil || res.StatusCode != http.StatusOK {
		log.Printf("%s is NOT healthy.", lb.servers[idx].addr)
		lb.servers[idx].healthy = false
	} else {
		lb.servers[idx].healthy = true
	}
	return nil
}
