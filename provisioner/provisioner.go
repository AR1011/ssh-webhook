package provisioner

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/AR1011/ssh-webhook/store"
	"github.com/AR1011/ssh-webhook/types"
)

type Provisioner struct {
	PublicURL   string
	InternalURL string
	store       store.Store
}

func NewProvisioner(store store.Store) *Provisioner {
	return &Provisioner{
		store: store,
	}
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
	// random port between 20000 - 65000
	rand.Seed(time.Now().UnixNano())

	// Define the range for the random port
	min := 20000
	max := 65000

	// Generate a random number within the range
	randomPort := min + rand.Intn(max-min+1)

	return int64(randomPort)
}

func (p *Provisioner) GetHookConfig(urli string) (types.WebhookConfig, error) {
	parsedUrl, err := url.Parse(urli)
	if err != nil {
		return types.WebhookConfig{}, err
	}

	hostParts := strings.Split(parsedUrl.Host, ":")
	host := hostParts[0]
	if host == "localhost" {
		host = "127.0.0.1"
	}

	port := int64(80)
	if len(hostParts) > 1 {
		port = int64(port)
	}

	clientSocket := types.Socket{
		Host: host,
		Port: port,
	}

	internalSocket := p.ProvisionSocket()
	publicURL := fmt.Sprintf("https://%s%s", parsedUrl.Host, parsedUrl.Path)

	return types.WebhookConfig{
		ClientSocket:         clientSocket,
		InternalServerSocket: internalSocket,
		Path:                 parsedUrl.Path,
		PublicURL:            publicURL,
		InternalURL:          fmt.Sprintf("http://%s%s", internalSocket.Socket(), parsedUrl.Path),
	}, nil
}
