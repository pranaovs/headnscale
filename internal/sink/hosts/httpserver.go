package hosts

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type httpServer struct {
	ip     net.IP
	port   int
	server *http.Server
	mux    *http.ServeMux
}

func newHTTPServer(ip net.IP, port int) *httpServer {
	return &httpServer{
		ip:   ip,
		port: port,
		mux:  http.NewServeMux(),
	}
}

func (h *httpServer) serve(path string, filePath string) {
	h.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	})
	log.Printf("Registered HTTP endpoint: %s -> %s", path, filePath)
}

func (h *httpServer) start(ctx context.Context) error {
	var addr string
	if h.ip.To4() == nil {
		addr = fmt.Sprintf("[%s]:%d", h.ip.String(), h.port)
	} else {
		addr = fmt.Sprintf("%s:%d", h.ip.String(), h.port)
	}

	h.server = &http.Server{
		Addr:              addr,
		Handler:           h.mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

func (h *httpServer) stop() error {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return h.server.Shutdown(ctx)
	}
	return nil
}
