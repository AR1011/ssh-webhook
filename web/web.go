package web

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/AR1011/ssh-webhook/provisioner"
)

type WebServer struct {
	Host        string
	router      *http.ServeMux
	provisioner *provisioner.Provisioner
}

func NewWebServer(host string, provisioner *provisioner.Provisioner) *WebServer {
	return &WebServer{
		Host:        host,
		router:      http.NewServeMux(),
		provisioner: provisioner,
	}
}

func (s *WebServer) createRoutes() {
	s.router.HandleFunc("/", s.handleRoot)
	s.router.HandleFunc("/{id}/*", s.handleID)
}

func (s *WebServer) Start() {
	s.createRoutes()
	fmt.Printf("Starting server on \t%s\n", s.Host)
	log.Fatal(http.ListenAndServe(s.Host, s.router))
}

func (s *WebServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func (s *WebServer) handleID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	fwdadr, err := s.provisioner.GetForwardingAddress(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	nr := http.Request{
		Method: r.Method,
		URL:    fwdadr,
		Header: r.Header,
		Body:   r.Body,
	}

	resp, err := http.DefaultClient.Do(&nr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	resp.Body.Close()

	return

}
