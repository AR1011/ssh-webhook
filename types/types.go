package types

import "fmt"

type InternalRoute struct {
	ID      string
	Adr     string
	ReqPath string
	Rlimit  int
	Rcount  int
	Timeout int
}

type Socket struct {
	Host string
	Port int64
}

func (s *Socket) Socket() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type WebhookConfig struct {
	ClientSocket         Socket
	InternalServerSocket Socket
	Path                 string
	PublicURL            string
	InternalURL          string
}

func (w WebhookConfig) String() string {
	return fmt.Sprintf("ClientSocket: %s\nInternalServerSocket: %s\nPublicURL: %s\nInternalURL: %s\n", w.ClientSocket.Socket(), w.InternalServerSocket.Socket(), w.PublicURL, w.InternalURL)
}
