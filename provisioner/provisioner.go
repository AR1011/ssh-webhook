package provisioner

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"

	"github.com/AR1011/ssh-webhook/store"
	"github.com/AR1011/ssh-webhook/types"
	"github.com/google/uuid"
)

type Provisioner struct {
	PublicURL   string
	InternalURL string
	Store       store.Store
	Dev         bool
}

func (p *Provisioner) GetForwardingAddress(id string) (*url.URL, error) {
	session, err := p.Store.GetByID(id)
	if err != nil {
		return nil, err
	}

	fwdURL, err := url.Parse(session.InternalURL)
	if err != nil {
		return nil, err
	}

	return fwdURL, nil
}

func (p *Provisioner) GetConfig(id string) (types.WebhookConfig, error) {
	session, err := p.Store.GetByID(id)
	if err != nil {
		return types.WebhookConfig{}, err
	}

	return session, nil
}

func (p *Provisioner) ProvisionSocket() types.Socket {
	return types.Socket{
		Host: "127.0.0.1",
		Port: p.RandomUnassignedPort(),
	}
}

func (p *Provisioner) RandomUnassignedPort() int64 {
	var (
		port int64
		min  = 20000
		max  = 65000
	)

	for {
		tp := min + rand.Intn(max-min+1)
		if p.Store.IsPortAvailable(tp) {
			port = int64(tp)
			break
		}
	}

	return port
}

func (p *Provisioner) GetHookConfig(urli string) (types.WebhookConfig, error) {
	id := uuid.New().String()

	parsedUrl, err := url.Parse(urli)
	if err != nil {
		return types.WebhookConfig{}, err
	}

	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return types.WebhookConfig{}, fmt.Errorf("invalid scheme: %s - must be http or https", parsedUrl.Scheme)
	}

	hostParts := strings.Split(parsedUrl.Host, ":")
	host := hostParts[0]
	if host == "localhost" {
		host = "127.0.0.1"
	}
	if host == "" {
		return types.WebhookConfig{}, fmt.Errorf("invalid host")
	}

	port := int64(80)
	if parsedUrl.Scheme == "https" {
		port = int64(443)
	}

	if len(hostParts) > 1 {
		var err error
		port, err = strconv.ParseInt(hostParts[1], 10, 64)
		if err != nil {
			return types.WebhookConfig{}, fmt.Errorf("invalid port: %v", err)
		}
	}

	clientSocket := types.Socket{
		Host: host,
		Port: port,
	}

	internalSocket := p.ProvisionSocket()

	var publicURL string
	if p.Dev {
		publicURL = fmt.Sprintf("%s/%s", p.InternalURL, id)
	} else {
		publicURL = fmt.Sprintf("%s/%s", p.PublicURL, id)
	}

	path := parsedUrl.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	internalURL := fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, internalSocket.Socket(), path)

	return types.WebhookConfig{
		ID:                   id,
		ClientSocket:         clientSocket,
		InternalServerSocket: internalSocket,
		Path:                 path,
		PublicURL:            publicURL,
		InternalURL:          internalURL,
	}, nil
}
