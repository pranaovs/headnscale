package hosts

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// startServer configures, listens, and serves the hosts file on a specific IP/Port.
// It returns the running *http.Server instance or an error.
func startServer(ip net.IP, port int, filePath string) (*http.Server, error) {
	mux := http.NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}

	mux.HandleFunc("/hosts", handler)
	mux.HandleFunc("/hosts.txt", handler)

	// 2. Determine Network (Explicitly separate IPv4 and IPv6)
	network := "tcp4"
	addr := fmt.Sprintf("%s:%d", ip.String(), port)

	if ip.To4() == nil {
		network = "tcp6"
		addr = fmt.Sprintf("[%s]:%d", ip.String(), port)
	}

	// 3. Create Listener
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// 4. Create Server
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 5. Start in Background
	go func() {
		log.Printf("Starting HTTP server on %s (%s)", addr, network)
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error (%s): %v", addr, err)
		}
	}()

	return srv, nil
}
