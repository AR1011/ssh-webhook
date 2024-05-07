package types

import (
	"fmt"
	"time"

	gssh "github.com/gliderlabs/ssh"
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
	Session   gssh.Session
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
	return fmt.Sprintf("ssh -R 0:%s localhost -p 2222 tunnel", w.ClientSocket.Socket())
}

func ToSizeUnit(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(n)/1024)
	}
	if n < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(n)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(n)/(1024*1024*1024))
}

type RequestAnalytic struct {
	Method           string
	ReceivedAt       time.Time
	TimeTaken        time.Duration
	From             string
	RequestBodySize  int64
	ResponseBodySize int64
	ResponseCode     int
}

func (r RequestAnalytic) String() string {
	return fmt.Sprintf(
		"Method=%s | ReceivedAt=%s | TimeTaken=%s | From=%s | RequestBodySize=%s | ResponseBodySize=%s | ResponseCode=%d",
		r.Method,
		r.ReceivedAt.Format("2006-01-02 15:04:05"),
		r.TimeTaken.Round(time.Millisecond),
		r.From,
		ToSizeUnit(r.RequestBodySize),
		ToSizeUnit(r.ResponseBodySize),
		r.ResponseCode,
	)
}
