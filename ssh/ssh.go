package ssh

import gossh "golang.org/x/crypto/ssh"

type SSHServer struct {
	Addr    string
	Handler func()
	Config  *gossh.ServerConfig
}

func SSHHandler(gossh.Session) {}
