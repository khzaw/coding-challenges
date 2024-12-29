package lb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type (
	server struct {
		addr   string
		status atomic.Bool
	}
	LB struct {
		port       int
		servers    []*server
		currentIdx atomic.Int32
		listener   net.Listener
		wg         sync.WaitGroup
		shutdown   chan struct{}
	}
)

func New(port int, serverArgs []string) *LB {
	lb := &LB{
		port:     port,
		shutdown: make(chan struct{}),
	}
	lb.currentIdx.Store(-1)

	var srvs []*server
	for _, a := range serverArgs {
		port, err := strconv.Atoi(a)
		if err != nil || port == lb.port || port < 0 || port > 65535 {
			log.Fatalf("invalid port: %v", err)
		}
		srv := &server{addr: "127.0.0.1:" + a}
		srv.status.Store(true)
		srvs = append(srvs, srv)
	}
	lb.servers = srvs

	return lb
}

func (lb *LB) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", lb.port))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	lb.listener = listener

	ticker := time.NewTicker(5 * time.Second)

	lb.wg.Add(1)
	go func() {
		defer lb.wg.Done()
		lb.runHealthChecks(ctx, ticker)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {

				if errors.Is(err, net.ErrClosed) {
					return nil
				}

				select {
				case <-ctx.Done():
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
		log.Printf("there is no healthy server")
		return
	}

	backendConn, err := net.Dial("tcp", nextServer.addr)
	if err != nil {
		log.Printf("failed to connect to BE server: %v\n", err)
		return
	}
	defer backendConn.Close()

	go io.Copy(backendConn, conn)
	go io.Copy(conn, backendConn)

	select {
	case <-lb.shutdown:
		return
	}
}

func (lb *LB) getNextHealthyServer() (*server, error) {
	for i := 0; i < len(lb.servers); i++ {
		current := lb.currentIdx.Load()
		next := (current + 1) % int32(len(lb.servers))

		if !lb.currentIdx.CompareAndSwap(current, next) {
			i--
			continue
		}

		if lb.servers[next].status.Load() {
			return lb.servers[next], nil
		}

	}

	return nil, errors.New("no healthy server")
}

func (lb *LB) runHealthChecks(ctx context.Context, ticker *time.Ticker) {
	defer ticker.Stop()

	pool := make(chan struct{}, runtime.GOMAXPROCS(0))

	for {
		select {
		case <-ticker.C:
			for idx := range lb.servers {
				pool <- struct{}{}
				go func(idx int) {
					lb.sendHealthCheck(ctx, idx)
					<-pool
				}(idx)
			}
		case <-ctx.Done():
			log.Println("Health checks stopped.")
			return
		}
	}
}

func (lb *LB) sendHealthCheck(ctx context.Context, idx int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+lb.servers[idx].addr+"/healthcheck", nil)
	if err != nil {
		log.Printf("failed to create a health check request")
		return
	}

	res, err := (&http.Client{}).Do(req)

	if err != nil {
		log.Printf("%s is NOT healthy.", lb.servers[idx].addr)
		lb.servers[idx].status.Store(false)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("%s is NOT healthy.", lb.servers[idx].addr)
		lb.servers[idx].status.Store(false)
	} else {
		lb.servers[idx].status.Store(true)
	}
	return
}
