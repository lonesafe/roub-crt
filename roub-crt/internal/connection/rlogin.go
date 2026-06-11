package connection

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type RloginConn struct {
	Conn    net.Conn
	Timeout time.Duration
}

type RloginConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	TermType string
	Timeout  time.Duration
}

func NewRloginConfig(host string, port int, username string) *RloginConfig {
	return &RloginConfig{
		Host:     host,
		Port:     port,
		Username: username,
		TermType: "vt100",
		Timeout:  30 * time.Second,
	}
}

func ConnectRlogin(config *RloginConfig) (*RloginConn, error) {
	if config.Port == 0 {
		config.Port = 513
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := net.DialTimeout("tcp", addr, config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	rlogin := &RloginConn{
		Conn:    conn,
		Timeout: config.Timeout,
	}

	if err := rlogin.handshake(config); err != nil {
		conn.Close()
		return nil, err
	}

	return rlogin, nil
}

func (r *RloginConn) handshake(config *RloginConfig) error {
	zero := byte(0)

	r.Conn.SetWriteDeadline(time.Now().Add(r.Timeout))

	r.Conn.Write([]byte{zero})
	r.Conn.Write([]byte(config.Username + "\x00"))
	r.Conn.Write([]byte(config.TermType + "\x00"))
	r.Conn.Write([]byte(config.Username + "\x00"))
	r.Conn.Write([]byte("vt100/9600\x00"))

	return nil
}

func (r *RloginConn) Read(p []byte) (n int, err error) {
	r.Conn.SetReadDeadline(time.Now().Add(r.Timeout))
	return r.Conn.Read(p)
}

func (r *RloginConn) Write(p []byte) (n int, err error) {
	r.Conn.SetWriteDeadline(time.Now().Add(r.Timeout))
	return r.Conn.Write(p)
}

func (r *RloginConn) Close() error {
	return r.Conn.Close()
}

func (r *RloginConn) HandleSLC() error {
	buf := make([]byte, 256)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		data := string(buf[:n])
		if strings.Contains(data, "BINARY") || strings.Contains(data, "\x00") {
			r.Write([]byte{0x01})
		}
	}
}
