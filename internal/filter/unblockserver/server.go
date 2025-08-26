package unblockserver

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	networkRules networkRules
	httpSrv      *http.Server
}

type networkRules interface {
	ParseRule(rule string, filterName *string) (isException bool, err error)
}

func NewServer(nr networkRules) *Server {
	return &Server{networkRules: nr}
}

func (s *Server) handleUnblock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	host := strings.ToLower(r.URL.Query().Get("host"))
	if host == "" {
		http.Error(w, "missing host", http.StatusBadRequest)
		return
	}

	rule := fmt.Sprintf("@@||%s^", host)

	filterList := "allowed-list"
	if _, err := s.networkRules.ParseRule(rule, &filterList); err != nil {
		http.Error(w, "networkrules: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("unblocked " + host))
}

func (s *Server) Start(port int) (int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/unblock", s.handleUnblock)

	s.httpSrv = &http.Server{
		Handler:      mux,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return -1, fmt.Errorf("listen: %w", err)
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port
	log.Printf("Unblock server listening on port %d", actualPort)

	go func() {
		if err := s.httpSrv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("error serving unblock: %v", err)
		}
	}()

	return actualPort, nil
}

func (s *Server) Stop() error {
	if s.httpSrv != nil {
		return s.httpSrv.Close()
	}
	return nil
}
