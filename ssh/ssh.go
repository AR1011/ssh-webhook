package ssh

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/AR1011/ssh-webhook/provisioner"
	gssh "github.com/gliderlabs/ssh"
	cssh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func HelpMessage() string {
	return `
SSH Wehook Server

Commands:
  setup 	Setup a new webhook
  tunnel 	Get tunnel config for existing webhook
  list 		List all webhooks
  active 	List all active ssh tunnels
  help 		Show this message

`
}

func startMessage() string {
	return `
                       _     _                 _    
                      | |   | |               | |   
  ___ _____      _____| |__ | |__   ___   ___ | | __
 / __/ __\ \ /\ / / _ \ '_ \| '_ \ / _ \ / _ \| |/ /
 \__ \__ \\ V  V /  __/ |_) | | | | (_) | (_) |   < 
 |___/___/ \_/\_/ \___|_.__/|_| |_|\___/ \___/|_|\_\
   
       

Commands:
  setup 	Setup a new webhook
  tunnel 	Get tunnel config for existing webhook
  list 		List all webhooks
  active 	List all active ssh tunnels
  help 		Show this message

`
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

type SSHServer struct {
	Server      *gssh.Server
	provisioner *provisioner.Provisioner
}

func NewSSHServer(addr string, provisioner *provisioner.Provisioner) *SSHServer {

	// using custome handler
	forwardHandler := &ForwardedTCPHandler{
		Provisioner: provisioner,
	}

	sshServer := gssh.Server{
		Addr: addr,
		ServerConfigCallback: func(ctx gssh.Context) *cssh.ServerConfig {
			cfg := &cssh.ServerConfig{
				ServerVersion: "SSH-2.0-sshhook",
			}

			cfg.Ciphers = []string{"aes128-gcm@openssh.com"}
			return cfg

		},
		PublicKeyHandler: func(ctx gssh.Context, key gssh.PublicKey) bool {

			// store the public_key in the context as we can only access that here
			ctx.SetValue("public_key", key.Marshal())
			return true
		},
		RequestHandlers: map[string]gssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
	}

	server := &SSHServer{
		Server:      &sshServer,
		provisioner: provisioner,
	}

	server.Server.Handle(func(s gssh.Session) {
		server.handleSession(s)
	})

	pk, err := server.readPK()
	if err != nil {
		log.Fatalf("failed to read private key: %s", err.Error())
	}

	server.Server.AddHostKey(pk)

	return server
}

func (s *SSHServer) Start() error {
	fmt.Printf("Starting SSH server on \t%s\n", s.Server.Addr)
	return s.Server.ListenAndServe()
}

func (s *SSHServer) readPK() (cssh.Signer, error) {
	b, err := os.ReadFile("keys/hostkey")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %s", err.Error())
	}

	pk, err := cssh.ParsePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("faailed to parse private key: %s", err.Error())
	}

	return pk, nil
}

func (s *SSHServer) handleTunnelSession(session gssh.Session) {

	webhookCondig, err := s.provisioner.Store.GetBySessionID(session.Context().SessionID())
	if err != nil {
		session.Write([]byte("Error: " + err.Error()))
		return
	}
	webhookCondig.ActiveSession.Session = session

	session.Write([]byte(fmt.Sprintf("Tunnel Conected\n\nPublic URL:\n%s\n\nLocal Binding:\n%s\n\nListening For Requests...\n\n", webhookCondig.PublicURL, webhookCondig.ClientSocket.Socket()+webhookCondig.Path)))

	// cleanup
	defer func() {
		webhookCondig.ActiveSession = nil
	}()

	<-session.Context().Done()
}

func (s *SSHServer) handleSession(session gssh.Session) {
	defer session.Close()

	// check the context for cancel session
	if session.Context().Value("close_session") != nil {
		session.Write([]byte(session.Context().Value("close_session_msg").(string)))
		return
	}

	if session.RawCommand() != "" {
		s.handleTunnelSession(session)
		return
	}

	t := term.NewTerminal(session, "$ ")
	t.Write([]byte(startMessage()))

	for {
		cmd, err := t.ReadLine()
		if err != nil {
			if err == io.EOF {
				t.Write([]byte("Goodbye\n"))
				return
			}
			t.Write([]byte("error reading from session: " + err.Error()))

			continue
		}

		t.Write([]byte("\033[H\033[2J"))

		switch cmd {
		case "setup":
			s.setupWebhook(session, t)
		case "tunnel":
			s.setupTunnel(session, t)
		case "list":
			s.listWebhooks(session)
		case "active":
			s.listActiveTunnels(session)
		case "help":
			s.help(session)
		case "exit":
			t.Write([]byte("Goodbye\n"))
			return
		default:
			t.Write([]byte(fmt.Sprintf("\nInvalid command\n%s", HelpMessage())))
		}
	}
}

func (s *SSHServer) renderMenu(t *term.Terminal) {
	t.Write([]byte("\033[H\033[2J"))
	t.Write([]byte(startMessage()))
}

func (s *SSHServer) setupWebhook(session gssh.Session, t *term.Terminal) {
	if session.PublicKey() == nil {
		t.Write([]byte("missing public key\n"))
		session.Close()
		return
	}

	t.SetPrompt("> ")
	t.Write([]byte("Welcome to ssh-hook\n\nEnter the Local URL to forward to:\n"))

	var url string
	var err error
	for {
		url, err = t.ReadLine()
		if err != nil {
			t.Write([]byte(fmt.Sprintf("\nError reading from session: %s\n", err.Error())))
			session.Close()
			return
		}

		if url == "exit" {

			s.renderMenu(t)
			return
		}

		// if no protocol, add http
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}

		whConfig, err := s.provisioner.GetHookConfig(url)
		if err != nil {
			t.Write([]byte(fmt.Sprintf("\nInvalid URL, please try again: %s\n", err.Error())))
			continue
		}

		key := session.PublicKey().Marshal()
		whConfig.PubKey = base64.StdEncoding.EncodeToString(key)
		s.provisioner.Store.Set(whConfig.ID, whConfig)

		fmt.Println(whConfig.String())

		t.Write([]byte(fmt.Sprintf("\n\nPublic URL: \n%s\n\nTunnel Command: \n%s\n\n", whConfig.PublicURL, whConfig.TunnelCommand())))
		session.Close()
		break
	}

}

func (s *SSHServer) setupTunnel(session gssh.Session, t *term.Terminal) {
	t.SetPrompt("> ")
	t.Write([]byte("Enter the webhook ID: "))
	id, err := t.ReadLine()
	if err != nil {
		t.Write([]byte("error reading from session: " + err.Error()))
		session.Close()
		return
	}

	t.Write([]byte("Enter the Local Port to forward to: "))
	port, err := t.ReadLine()
	if err != nil {
		t.Write([]byte("error reading from session: " + err.Error()))
		session.Close()
		return
	}

	fmt.Printf("ID: %s, Port: %s\n", id, port)

}

func (s *SSHServer) listWebhooks(session gssh.Session) {
	session.Write([]byte("List of webhooks\n"))
}

func (s *SSHServer) listActiveTunnels(session gssh.Session) {
	session.Write([]byte("List of active tunnels\n"))
}

func (s *SSHServer) help(session gssh.Session) {
	session.Write([]byte(HelpMessage()))
}
