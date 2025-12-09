package httpserver

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func Start(ip net.IP, port int) {
	var addr string

	// IPv6 must use [ip]:port
	if ip.To4() == nil {
		addr = fmt.Sprintf("[%s]:%d", ip.String(), port)
	} else {
		addr = fmt.Sprintf("%s:%d", ip.String(), port)
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func ServeFile(path string, filePath string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	})
}
