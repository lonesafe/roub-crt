package connection

import (
	"fmt"
	"io"
	"sync"

	"go.bug.st/serial"
)

type SerialConn struct {
	Port   serial.Port
	RW     io.ReadWriter
	mu     sync.Mutex
	isOpen bool
}

type SerialConfig struct {
	PortName string
	BaudRate int
	DataBits int
	StopBits serial.StopBits
	Parity   serial.Parity
}

func NewSerialConfig(portName string, baudRate int) *SerialConfig {
	return &SerialConfig{
		PortName: portName,
		BaudRate: baudRate,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}
}

func ConnectSerial(config *SerialConfig) (*SerialConn, error) {
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: config.DataBits,
		StopBits: config.StopBits,
		Parity:   config.Parity,
	}

	t, err := serial.Open(config.PortName, mode)
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

func (s *SerialConn) SetBaudRate(baud int) error {
	return s.Port.SetMode(&serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	})
}

func ListSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to list serial ports: %w", err)
	}
	return ports, nil
}
