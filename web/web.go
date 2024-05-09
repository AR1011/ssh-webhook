package web

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/AR1011/ssh-webhook/provisioner"
	"github.com/AR1011/ssh-webhook/types"
)

type WebServer struct {
	Host        string
	router      *http.ServeMux
	provisioner *provisioner.Provisioner
}

func NewWebServer(host string, provisioner *provisioner.Provisioner) *WebServer {
	return &WebServer{
		Host:        host,
		router:      nil,
		provisioner: provisioner,
	}
}

func (s *WebServer) createRoutes() {
	s.router = http.NewServeMux()
	s.router.HandleFunc("/{id}", s.handleID)
	s.router.HandleFunc("/{id}/*", s.handleID)
}

func (s *WebServer) Start() {
	s.createRoutes()
	fmt.Printf("Starting server on \t%s\n", s.Host)
	log.Fatal(http.ListenAndServe(s.Host, s.router))
}

func (s *WebServer) handleID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	st := time.Now()
	reqAnalytic := types.RequestAnalytic{
		Method:     r.Method,
		From:       r.RemoteAddr,
		ReceivedAt: time.Now(),
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	reqAnalytic.RequestBodySize = int64(len(b))

	defer r.Body.Close()

	condig, err := s.provisioner.GetConfig(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fwdAddr, err := url.Parse(condig.InternalURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	nr, err := http.NewRequest(r.Method, fwdAddr.String(), bytes.NewReader(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nr.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(nr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	l, err := io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reqAnalytic.ResponseBodySize = l
	reqAnalytic.ResponseCode = resp.StatusCode
	reqAnalytic.TimeTaken = time.Since(st)

	if condig.ActiveSession.Session != nil {
		condig.ActiveSession.Session.Write([]byte(fmt.Sprintf("\n%s\n", reqAnalytic.String())))
	}

	fmt.Println(reqAnalytic.String())
}
