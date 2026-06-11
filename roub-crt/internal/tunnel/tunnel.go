package tunnel

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type TunnelType string

const (
	TunnelLocal  TunnelType = "local"
	TunnelRemote TunnelType = "remote"
	TunnelDynamic TunnelType = "dynamic"
)

type Tunnel struct {
	Type       TunnelType
	ListenAddr string
	ListenPort int
	TargetHost string
	TargetPort int
	SocksPort  int
	Client     *ssh.Client
	Listener   net.Listener
	Done       chan struct{}
	mu         sync.Mutex
	isClosed   bool
}

type TunnelManager struct {
	tunnels map[string]*Tunnel
	mu      sync.RWMutex
}

func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		tunnels: make(map[string]*Tunnel),
	}
}

func (tm *TunnelManager) CreateLocalTunnel(client *ssh.Client, listenPort int, targetHost string, targetPort int) (*Tunnel, error) {
	addr := fmt.Sprintf("localhost:%d", listenPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	tunnel := &Tunnel{
		Type:       TunnelLocal,
		ListenAddr: "localhost",
		ListenPort: listenPort,
		TargetHost: targetHost,
		TargetPort: targetPort,
		Client:     client,
		Listener:   listener,
		Done:       make(chan struct{}),
	}

	go tunnel.handleLocalConnections()

	tm.mu.Lock()
	tm.tunnels[tunnel.addr()] = tunnel
	tm.mu.Unlock()

	return tunnel, nil
}

func (t *Tunnel) addr() string {
	return fmt.Sprintf("%s:%d", t.ListenAddr, t.ListenPort)
}

func (t *Tunnel) handleLocalConnections() {
	defer t.Listener.Close()

	for {
		select {
		case <-t.Done:
			return
		default:
		}

		t.Listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := t.Listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		go t.forwardLocal(conn)
	}
}

func (t *Tunnel) forwardLocal(clientConn net.Conn) {
	defer clientConn.Close()

	targetAddr := fmt.Sprintf("%s:%d", t.TargetHost, t.TargetPort)
	serverConn, err := t.Client.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	defer serverConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(serverConn, clientConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(clientConn, serverConn)
	}()

	wg.Wait()
}

func (tm *TunnelManager) CreateRemoteTunnel(client *ssh.Client, bindAddr string, bindPort int, targetHost string, targetPort int) (*Tunnel, error) {
	addr := fmt.Sprintf("%s:%d", bindAddr, bindPort)

	listener, err := client.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote listener on %s: %w", addr, err)
	}

	tunnel := &Tunnel{
		Type:       TunnelRemote,
		ListenAddr: bindAddr,
		ListenPort: bindPort,
		TargetHost: targetHost,
		TargetPort: targetPort,
		Client:     client,
		Listener:   listener,
		Done:       make(chan struct{}),
	}

	go tunnel.handleRemoteConnections()

	tm.mu.Lock()
	tm.tunnels[tunnel.addr()] = tunnel
	tm.mu.Unlock()

	return tunnel, nil
}

func (t *Tunnel) handleRemoteConnections() {
	defer t.Listener.Close()

	for {
		select {
		case <-t.Done:
			return
		default:
		}

		conn, err := t.Listener.Accept()
		if err != nil {
			return
		}

		go t.forwardRemote(conn)
	}
}

func (t *Tunnel) forwardRemote(clientConn net.Conn) {
	defer clientConn.Close()

	targetAddr := fmt.Sprintf("%s:%d", t.TargetHost, t.TargetPort)
	serverConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	defer serverConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(serverConn, clientConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(clientConn, serverConn)
	}()

	wg.Wait()
}

func (tm *TunnelManager) CreateDynamicTunnel(client *ssh.Client, listenPort int) (*Tunnel, error) {
	addr := fmt.Sprintf("localhost:%d", listenPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	tunnel := &Tunnel{
		Type:      TunnelDynamic,
		ListenAddr: "localhost",
		ListenPort: listenPort,
		SocksPort:  listenPort,
		Client:     client,
		Listener:   listener,
		Done:       make(chan struct{}),
	}

	go tunnel.handleSOCKS()

	tm.mu.Lock()
	tm.tunnels[fmt.Sprintf("socks:%d", listenPort)] = tunnel
	tm.mu.Unlock()

	return tunnel, nil
}

func (t *Tunnel) handleSOCKS() {
	defer t.Listener.Close()

	for {
		select {
		case <-t.Done:
			return
		default:
		}

		t.Listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := t.Listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		go t.handleSOCKSConnection(conn)
	}
}

func (t *Tunnel) handleSOCKSConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	if buf[0] != 0x05 {
		return
	}

	conn.Write([]byte{0x05, 0x00})

	n, err = conn.Read(buf)
	if err != nil || n < 10 {
		return
	}

	if buf[0] != 0x05 || buf[1] != 0x01 {
		return
	}

	var targetHost string
	var targetPort int

	switch buf[3] {
	case 0x01:
		targetHost = net.IP(buf[4:8]).String()
		targetPort = int(buf[8])<<8 + int(buf[9])
	case 0x03:
		addrLen := int(buf[4])
		targetHost = string(buf[5 : 5+addrLen])
		targetPort = int(buf[5+addrLen])<<8 + int(buf[6+addrLen])
	case 0x04:
		targetHost = net.IP(buf[4:20]).String()
		targetPort = int(buf[20])<<8 + int(buf[21])
	default:
		return
	}

	targetConn, err := t.Client.Dial("tcp", fmt.Sprintf("%s:%d", targetHost, targetPort))
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}
	defer targetConn.Close()

	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(targetConn, conn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(conn, targetConn)
	}()

	wg.Wait()
}

func (tm *TunnelManager) CloseTunnel(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tunnel, ok := tm.tunnels[name]
	if !ok {
		return fmt.Errorf("tunnel not found: %s", name)
	}

	tunnel.mu.Lock()
	if !tunnel.isClosed {
		tunnel.isClosed = true
		close(tunnel.Done)
	}
	tunnel.mu.Unlock()

	if tunnel.Listener != nil {
		tunnel.Listener.Close()
	}

	delete(tm.tunnels, name)
	return nil
}

func (tm *TunnelManager) ListTunnels() []*Tunnel {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tunnels := make([]*Tunnel, 0, len(tm.tunnels))
	for _, t := range tm.tunnels {
		tunnels = append(tunnels, t)
	}
	return tunnels
}

func (tm *TunnelManager) CloseAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for name := range tm.tunnels {
		tm.CloseTunnel(name)
	}
}

type ForwardedTCPChannel struct {
	Channel  ssh.Channel
	Request  string
	Listener net.Listener
}

func (sf *Tunnel) String() string {
	switch sf.Type {
	case TunnelLocal:
		return fmt.Sprintf("local:%d -> %s:%d", sf.ListenPort, sf.TargetHost, sf.TargetPort)
	case TunnelRemote:
		return fmt.Sprintf("remote:%s:%d -> %s:%d", sf.ListenAddr, sf.ListenPort, sf.TargetHost, sf.TargetPort)
	case TunnelDynamic:
		return fmt.Sprintf("dynamic:socks5://localhost:%d", sf.SocksPort)
	default:
		return "unknown"
	}
}

type TunnelInfo struct {
	Name    string
	Type    TunnelType
	Local   string
	Remote  string
	Status  string
}

func (tm *TunnelManager) GetTunnelInfo() []TunnelInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	infos := make([]TunnelInfo, 0, len(tm.tunnels))
	for name, t := range tm.tunnels {
		info := TunnelInfo{
			Name:   name,
			Type:   t.Type,
			Status: "active",
		}

		switch t.Type {
		case TunnelLocal:
			info.Local = fmt.Sprintf("%s:%d", t.ListenAddr, t.ListenPort)
			info.Remote = fmt.Sprintf("%s:%d", t.TargetHost, t.TargetPort)
		case TunnelRemote:
			info.Local = fmt.Sprintf("%s:%d", t.ListenAddr, t.ListenPort)
			info.Remote = fmt.Sprintf("%s:%d", t.TargetHost, t.TargetPort)
		case TunnelDynamic:
			info.Local = fmt.Sprintf("socks5://%s:%d", t.ListenAddr, t.SocksPort)
		}

		t.mu.Lock()
		if t.isClosed {
			info.Status = "closed"
		}
		t.mu.Unlock()

		infos = append(infos, info)
	}

	return infos
}
