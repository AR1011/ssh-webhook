package ssh

// modified from
// https://github.com/gliderlabs/ssh/blob/master/tcpip.go

// modified to add custom logic for forwarding handlers
// so user can ssh with
// ssh -R 0:127.0.0.1:3000 localhost -p 2222
// rather than
// ssh -R 127.0.0.1:23432:127.0.0.1:3000 localhost -p 2222
// so the server binding is all handled here and the user
// does not see or interact with that

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/AR1011/ssh-webhook/provisioner"
	"github.com/AR1011/ssh-webhook/types"
	gssh "github.com/gliderlabs/ssh"
	cssh "golang.org/x/crypto/ssh"
)

const (
	forwardedTCPChannelType = "forwarded-tcpip"
)

// direct-tcpip data struct as specified in RFC4254, Section 7.2
type localForwardChannelData struct {
	DestAddr string
	DestPort uint32

	OriginAddr string
	OriginPort uint32
}

// DirectTCPIPHandler can be enabled by adding it to the server's
// ChannelHandlers under direct-tcpip.
func DirectTCPIPHandler(srv *gssh.Server, conn *cssh.ServerConn, newChan cssh.NewChannel, ctx gssh.Context) {
	fmt.Println("Custom DirectTCPIPHandler")
	d := localForwardChannelData{}
	if err := cssh.Unmarshal(newChan.ExtraData(), &d); err != nil {
		newChan.Reject(cssh.ConnectionFailed, "error parsing forward data: "+err.Error())
		return
	}

	if srv.LocalPortForwardingCallback == nil || !srv.LocalPortForwardingCallback(ctx, d.DestAddr, d.DestPort) {
		newChan.Reject(cssh.Prohibited, "port forwarding is disabled")
		return
	}

	dest := net.JoinHostPort(d.DestAddr, strconv.FormatInt(int64(d.DestPort), 10))

	var dialer net.Dialer
	dconn, err := dialer.DialContext(ctx, "tcp", dest)
	if err != nil {
		newChan.Reject(cssh.ConnectionFailed, err.Error())
		return
	}

	ch, reqs, err := newChan.Accept()
	if err != nil {
		dconn.Close()
		return
	}
	go cssh.DiscardRequests(reqs)

	go func() {
		defer ch.Close()
		defer dconn.Close()
		io.Copy(ch, dconn)
	}()
	go func() {
		defer ch.Close()
		defer dconn.Close()
		io.Copy(dconn, ch)
	}()
}

type remoteForwardRequest struct {
	BindAddr string
	BindPort uint32
}

type remoteForwardSuccess struct {
	BindPort uint32
}

type remoteForwardCancelRequest struct {
	BindAddr string
	BindPort uint32
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

// ForwardedTCPHandler can be enabled by creating a ForwardedTCPHandler and
// adding the HandleSSHRequest callback to the server's RequestHandlers under
// tcpip-forward and cancel-tcpip-forward.
type ForwardedTCPHandler struct {
	forwards map[string]net.Listener
	sync.Mutex
	Provisioner *provisioner.Provisioner
}

// callback to set values in context, which will be checked later
// in the session to close the session
func closeSession(ctx gssh.Context) func(msg string) {
	return func(msg string) {
		ctx.SetValue("close_session", true)
		ctx.SetValue("close_session_msg", msg)
	}
}

func (h *ForwardedTCPHandler) HandleSSHRequest(ctx gssh.Context, srv *gssh.Server, req *cssh.Request) (bool, []byte) {
	fmt.Println("Custom ForwardedTCPHandler > HandleSSHRequest")
	fmt.Printf("ssh remote addr: \t%s\n", ctx.RemoteAddr().String())
	fmt.Printf("ssh local addr: \t%s\n", ctx.LocalAddr().String())

	h.Lock()
	if h.forwards == nil {
		h.forwards = make(map[string]net.Listener)
	}
	h.Unlock()
	conn := ctx.Value(gssh.ContextKeyConn).(*cssh.ServerConn)

	closeSessionCallback := closeSession(ctx)

	switch req.Type {
	case "tcpip-forward":

		fmt.Println("tcpip-forward")

		var reqPayload remoteForwardRequest
		if err := cssh.Unmarshal(req.Payload, &reqPayload); err != nil {
			closeSessionCallback("error parsing forward data: " + err.Error())
			return false, []byte{}
		}

		host := reqPayload.BindAddr
		port := reqPayload.BindPort

		// ensure the the request was sent like
		// ssh -R 0:<client-ip>:<client-port>
		if host != "localhost" || port != 0 {
			slog.Warn("attempt to bind", "host", host, "port", port, "error", "denied: invalid host or port")
			closeSessionCallback("invalid host or port")
			return false, []byte{}
		}

		fmt.Printf("attempt to bind %s:%d\n", host, port)

		// get public key from contect
		pk := ctx.Value("public_key")
		if pk == nil {
			slog.Warn("attempt to bind", "host", host, "port", port, "error", "denied: no public key")
			closeSessionCallback("no public key")
			return false, []byte{}
		}

		// ensure the key is bytes
		keyb, ok := pk.([]byte)
		if !ok {
			slog.Warn("attempt to bind", "host", host, "port", port, "error", "denied: no public key")
			closeSessionCallback("no public key")
			return false, []byte{}
		}

		// encode the key to string
		key, ok := base64.StdEncoding.EncodeToString(keyb), true
		if !ok {
			slog.Warn("attempt to bind", "host", host, "port", port, "error", "denied: no public key")
			closeSessionCallback("no public key")
			return false, []byte{}
		}

		// somehow identify how to get the config ??
		// we dont have:
		// session.RawCommand - to get arbitrary id
		// local target address eg (localhost:3000)
		//
		// we have:
		// public key
		// the new sessionID for the ssh -R 0:localhost:3000 session
		//
		config, err := h.Provisioner.Store.FindByPublicKeyAndSocket(key, host, int(port))
		if err != nil {
			slog.Warn("attempt to bind", "host", host, "port", port, "error", "denied: "+err.Error())
			closeSessionCallback(err.Error())
			return false, []byte{}
		}

		if config.ActiveSession != nil {
			slog.Warn("attempt to bind", "host", host, "port", port, "error", "denied: session already active")
			closeSessionCallback("session already active")
			return false, []byte{}
		}

		config.ActiveSession = &types.TunnelSession{
			SessionID: ctx.SessionID(),
			StartedAt: time.Now(),
		}
		h.Provisioner.Store.Update(config.ID, config)

		addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			// TODO: log listen failure
			return false, []byte{}
		}
		_, destPortStr, _ := net.SplitHostPort(ln.Addr().String())
		destPort, _ := strconv.Atoi(destPortStr)
		h.Lock()
		h.forwards[addr] = ln
		h.Unlock()
		go func() {
			<-ctx.Done()
			h.Lock()
			ln, ok := h.forwards[addr]
			h.Unlock()
			if ok {
				ln.Close()
			}
		}()
		go func() {
			for {
				c, err := ln.Accept()
				if err != nil {
					// TODO: log accept failure
					break
				}
				originAddr, orignPortStr, _ := net.SplitHostPort(c.RemoteAddr().String())
				originPort, _ := strconv.Atoi(orignPortStr)
				payload := cssh.Marshal(&remoteForwardChannelData{
					DestAddr:   reqPayload.BindAddr,
					DestPort:   uint32(destPort),
					OriginAddr: originAddr,
					OriginPort: uint32(originPort),
				})
				go func() {
					ch, reqs, err := conn.OpenChannel(forwardedTCPChannelType, payload)
					if err != nil {
						// TODO: log failure to open channel
						log.Println(err)
						c.Close()
						return
					}
					go cssh.DiscardRequests(reqs)
					go func() {
						defer ch.Close()
						defer c.Close()
						io.Copy(ch, c)
					}()
					go func() {
						defer ch.Close()
						defer c.Close()
						io.Copy(c, ch)
					}()
				}()
			}
			h.Lock()
			delete(h.forwards, addr)
			h.Unlock()
		}()
		return true, cssh.Marshal(&remoteForwardSuccess{uint32(destPort)})

	case "cancel-tcpip-forward":
		var reqPayload remoteForwardCancelRequest
		if err := cssh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		h.Lock()
		ln, ok := h.forwards[addr]
		h.Unlock()
		if ok {
			ln.Close()
		}
		return true, nil
	default:
		return false, nil
	}
}
