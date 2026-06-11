package connection

import (
	"fmt"
	"io"
	"sync"

	"github.com/pkg/term"
)

type SerialConn struct {
	Port    *term.Term
	RW      io.ReadWriter
	mu      sync.Mutex
	isOpen  bool
}

type SerialConfig struct {
	PortName string
	BaudRate int
	DataBits int
	StopBits int
	Parity   int
}

const (
	ParityNone = 0
	ParityOdd  = 1
	ParityEven = 2
)

func NewSerialConfig(portName string, baudRate int) *SerialConfig {
	return &SerialConfig{
		PortName: portName,
		BaudRate: baudRate,
		DataBits: 8,
		StopBits: 1,
		Parity:   ParityNone,
	}
}

func ConnectSerial(config *SerialConfig) (*SerialConn, error) {
	t, err := term.Open(config.PortName,
		term.Speed(config.BaudRate),
		term.RawMode,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", config.PortName, err)
	}

	return &SerialConn{
		Port:   t,
		RW:     t,
		isOpen: true,
	}, nil
}

func (s *SerialConn) Read(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.RW.Read(p)
}

func (s *SerialConn) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.RW.Write(p)
}

func (s *SerialConn) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isOpen {
		s.isOpen = false
		return s.Port.Close()
	}
	return nil
}

func (s *SerialConn) SetRawMode() error {
	return term.RawMode(s.Port)
}

func (s *SerialConn) SetBaudRate(baud int) error {
	return term.Speed(baud)(s.Port)
}

func ListSerialPorts() ([]string, error) {
	ports := []string{
		"/dev/ttyUSB0",
		"/dev/ttyUSB1",
		"/dev/ttyUSB2",
		"/dev/ttyUSB3",
		"/dev/ttyS0",
		"/dev/ttyS1",
		"/dev/ttyS2",
		"/dev/ttyS3",
		"/dev/ttyACM0",
		"/dev/ttyACM1",
		"COM1",
		"COM2",
		"COM3",
		"COM4",
	}

	var available []string
	for _, port := range ports {
		t, err := term.Open(port, term.Speed(9600))
		if err == nil {
			t.Close()
			available = append(available, port)
		}
	}

	return available, nil
}
