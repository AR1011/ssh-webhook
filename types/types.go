package types

import (
	"fmt"
)

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
	ID                   string
	ClientSocket         Socket
	InternalServerSocket Socket
	Path                 string
	PublicURL            string
	InternalURL          string
}

func (w WebhookConfig) String() string {
	return fmt.Sprintf("ClientSocket: \t%s\nServerSocket: \t%s\nPublicURL: \t%s\nInternalURL: \t%s\n", w.ClientSocket.Socket(), w.InternalServerSocket.Socket(), w.PublicURL, w.InternalURL)
}

func (w WebhookConfig) TunnelCommand() string {
	return fmt.Sprintf("ssh -R %s:%s localhost -p 2222", w.InternalServerSocket.Socket(), w.ClientSocket.Socket())
}
