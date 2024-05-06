package store

import "github.com/AR1011/ssh-webhook/types"

type Store interface {
	GetByID(id string) (types.WebhookConfig, error)
	GetBySessionID(id string) (types.WebhookConfig, error)
	Set(id string, value types.WebhookConfig) error
	Update(id string, value types.WebhookConfig) error
	Delete(id string) error
	IsPortAvailable(port int) bool
	FindByPublicKeyAndSocket(publicKey string, host string, port int) (types.WebhookConfig, error)
}
