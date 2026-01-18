package hosts

import (
	"log"
	"net"
	"net/http"
	"time"
)

// startHTTP configures and starts an HTTP server using the provided listener.
// It returns the *http.Server instance.
func startHTTP(ln net.Listener, filePath string) *http.Server {
	mux := http.NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}

	mux.HandleFunc("/", handler)

	// Create Server
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start in Background
	go func() {
		log.Printf("Starting HTTP server on %s (%s)", ln.Addr().String(), ln.Addr().Network())
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error (%s): %v", ln.Addr().String(), err)
		}
	}()

	return srv
}
