package example

import (
	"fmt"
	"io"
)

//go:generate go run ../main.go -type=MyService

// MyService is a concrete type with various methods
type MyService struct {
	name    string
	counter int
}

// NewMyService creates a new service instance
func NewMyService(name string) *MyService {
	return &MyService{
		name:    name,
		counter: 0,
	}
}

// GetName returns the service name
func (s *MyService) GetName() string {
	return s.name
}

// SetName sets the service name
func (s *MyService) SetName(name string) {
	s.name = name
}

// Increment increases the counter and returns the new value
func (s *MyService) Increment() int {
	s.counter++
	return s.counter
}

// Process handles data processing
func (s *MyService) Process(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	return append([]byte(s.name+": "), data...), nil
}

// Write implements io.Writer
func (s *MyService) Write(p []byte) (n int, err error) {
	fmt.Printf("[%s] %s", s.name, string(p))
	return len(p), nil
}

// ReadFrom reads from a reader
func (s *MyService) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 1024)
	total := int64(0)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			s.Write(buf[:n])
			total += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, err
		}
	}
}

// unexportedMethod is not exported so won't be in the interface
func (s *MyService) unexportedMethod() {
	// private implementation
}