package main

import (
	"fmt"

	"github.com/AR1011/ssh-webhook/provisioner"
	"github.com/AR1011/ssh-webhook/ssh"
	"github.com/AR1011/ssh-webhook/store"
	"github.com/AR1011/ssh-webhook/web"
)

func main() {
	fmt.Println(banner())
	store := store.NewMemoryStore()
	provisioner := provisioner.NewProvisioner(store)
	httpServer := web.NewWebServer(":4001", provisioner)
	sshServer := ssh.NewSSHServer(":2222", provisioner)

	go sshServer.Start()
	httpServer.Start()

}

func banner() string {
	return `
                       _     _                 _    
                      | |   | |               | |   
  ___ _____      _____| |__ | |__   ___   ___ | | __
 / __/ __\ \ /\ / / _ \ '_ \| '_ \ / _ \ / _ \| |/ /
 \__ \__ \\ V  V /  __/ |_) | | | | (_) | (_) |   < 
 |___/___/ \_/\_/ \___|_.__/|_| |_|\___/ \___/|_|\_\
                                                                                                        
`
}
