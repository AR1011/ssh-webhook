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
}

func (p *Provisioner) Provision(ir types.InternalRoute) error {
	return nil
}

func (p *Provisioner) Deprovision(id string) error {
	return nil
}

func (p *Provisioner) GetForwardingAddress(id string) (*url.URL, error) {
	return &url.URL{}, nil
}

func (p *Provisioner) ProvisionSocket() types.Socket {
	return types.Socket{
		Host: "127.0.0.1",
		Port: p.RandomUnassignedPort(),
	}
}

func (p *Provisioner) RandomUnassignedPort() int64 {
	min := 20000
	max := 65000
	randomPort := min + rand.Intn(max-min+1)

	return int64(randomPort)
}

func (p *Provisioner) GetHookConfig(urli string) (types.WebhookConfig, error) {
	fmt.Println("URL: ", urli)
	id := uuid.New().String()
	parsedUrl, err := url.Parse(urli)
	if err != nil {
		return types.WebhookConfig{}, err
	}

	fmt.Printf("Parsed URL: %+v\n", parsedUrl)
	hostParts := strings.Split(parsedUrl.Host, ":")
	host := hostParts[0]
	if host == "localhost" {
		host = "127.0.0.1"
	}
	if host == "" {
		return types.WebhookConfig{}, fmt.Errorf("invalid host")
	}

	port := int64(80)
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
	publicURL := fmt.Sprintf("%s/%s", p.PublicURL, id)

	return types.WebhookConfig{
		ID:                   id,
		ClientSocket:         clientSocket,
		InternalServerSocket: internalSocket,
		Path:                 parsedUrl.Path,
		PublicURL:            publicURL,
		InternalURL:          fmt.Sprintf("http://%s/%s", internalSocket.Socket(), parsedUrl.Path),
	}, nil
}
