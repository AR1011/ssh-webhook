package types

import (
	"fmt"
	"time"
)

type Socket struct {
	Host string
	Port int64
}

func (s *Socket) Socket() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type TunnelSession struct {
	SessionID string
	StartedAt time.Time
}

type WebhookConfig struct {
	PubKey               string
	ID                   string
	ClientSocket         Socket
	InternalServerSocket Socket
	Path                 string
	PublicURL            string
	InternalURL          string
	ActiveSession        *TunnelSession
}

func (w WebhookConfig) String() string {
	return fmt.Sprintf("PubKey: %s\n\nClientSocket: \t%s\nServerSocket: \t%s\nPublicURL: \t%s\nInternalURL: \t%s\n", w.PubKey, w.ClientSocket.Socket(), w.InternalServerSocket.Socket(), w.PublicURL, w.InternalURL)
}

func (w WebhookConfig) TunnelCommand() string {
	return fmt.Sprintf("ssh -R %s:%s localhost -p 2222 tunnel", w.InternalServerSocket.Socket(), w.ClientSocket.Socket())
}
