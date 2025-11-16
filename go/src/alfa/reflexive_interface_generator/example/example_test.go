package example

import (
	"bytes"
	"io"
	"testing"
)

// TestInterfaceGeneration tests that the generated interface works
func TestInterfaceGeneration(t *testing.T) {
	// Create a concrete instance
	service := NewMyService("test")

	// The generated interface should be assignable
	var _ IMyService = service

	// Test through interface
	var iface IMyService = service

	// Test GetName
	if name := iface.GetName(); name != "test" {
		t.Errorf("GetName() = %v, want %v", name, "test")
	}

	// Test SetName
	iface.SetName("updated")
	if name := iface.GetName(); name != "updated" {
		t.Errorf("After SetName, GetName() = %v, want %v", name, "updated")
	}

	// Test Increment
	if val := iface.Increment(); val != 1 {
		t.Errorf("Increment() = %v, want %v", val, 1)
	}
	if val := iface.Increment(); val != 2 {
		t.Errorf("Increment() = %v, want %v", val, 2)
	}

	// Test Process
	data := []byte("hello")
	result, err := iface.Process(data)
	if err != nil {
		t.Errorf("Process() error = %v", err)
	}
	expected := "updated: hello"
	if string(result) != expected {
		t.Errorf("Process() = %v, want %v", string(result), expected)
	}

	// Test Write (io.Writer)
	n, err := iface.Write([]byte("test data"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 9 {
		t.Errorf("Write() n = %v, want %v", n, 9)
	}

	// Test ReadFrom
	reader := bytes.NewReader([]byte("input data"))
	total, err := iface.ReadFrom(reader)
	if err != nil {
		t.Errorf("ReadFrom() error = %v", err)
	}
	if total != 10 {
		t.Errorf("ReadFrom() total = %v, want %v", total, 10)
	}
}

// MockService demonstrates implementing the generated interface
type MockService struct {
	name string
}

func (m *MockService) GetName() string {
	return m.name
}

func (m *MockService) SetName(name string) {
	m.name = name
}

func (m *MockService) Increment() int {
	return 42
}

func (m *MockService) Process(data []byte) ([]byte, error) {
	return []byte("mock: " + string(data)), nil
}

func (m *MockService) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *MockService) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}

// TestMockImplementation tests that mocks can implement the generated interface
func TestMockImplementation(t *testing.T) {
	mock := &MockService{name: "mock"}

	// Mock should implement the generated interface
	var iface IMyService = mock

	if name := iface.GetName(); name != "mock" {
		t.Errorf("Mock GetName() = %v, want %v", name, "mock")
	}

	if val := iface.Increment(); val != 42 {
		t.Errorf("Mock Increment() = %v, want %v", val, 42)
	}
}
