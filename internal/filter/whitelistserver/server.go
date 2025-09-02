package whitelistserver

import (
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"time"
)

type Server struct {
	networkRules networkRules
	httpSrv      *http.Server
	Port         int
}

type networkRules interface {
	ParseRule(rule string, filterName *string) (isException bool, err error)
}

func New(nr networkRules) *Server {
	return &Server{networkRules: nr}
}

func (s *Server) handleAllow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	rule := r.FormValue("rule")
	if rule == "" {
		http.Error(w, "missing rule", http.StatusBadRequest)
		return
	}

	filterList := "allowed-list"
	if _, err := s.networkRules.ParseRule(fmt.Sprintf("@@%s", rule), &filterList); err != nil {
		http.Error(w, "networkrules: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("allowed rule: " + html.EscapeString(rule)))
}

func (s *Server) Start() (int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/allow-rule", s.handleAllow)

	s.httpSrv = &http.Server{
		Handler:      mux,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	addr := fmt.Sprintf("127.0.0.1:%d", 0) // random port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return -1, fmt.Errorf("listen: %w", err)
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port
	log.Printf("Whitelist server listening on port %d", actualPort)

	go func() {
		if err := s.httpSrv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("error serving whitelist server: %v", err)
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
