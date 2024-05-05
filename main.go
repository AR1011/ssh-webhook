package main

import (
	"github.com/AR1011/ssh-webhook/provisioner"
	"github.com/AR1011/ssh-webhook/ssh"
	"github.com/AR1011/ssh-webhook/store"
	"github.com/AR1011/ssh-webhook/web"
)

func main() {
	store := store.NewMemoryStore()
	provisioner := &provisioner.Provisioner{
		PublicURL:   "https://sshwebhook.io",
		InternalURL: "http://127.0.0.1",
		Store:       store,
	}

	httpServer := web.NewWebServer(":4001", provisioner)
	sshServer := ssh.NewSSHServer(":2222", provisioner)

	go sshServer.Start()
	httpServer.Start()

}
