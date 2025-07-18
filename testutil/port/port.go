package port

import (
	"fmt"
	"net"
	"strconv"
)

// GetFreePort returns a free port number that can be used for testing
func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// MustGetFreePort is like GetFreePort but panics on error
func MustGetFreePort() int {
	port, err := GetFreePort()
	if err != nil {
		panic(fmt.Sprintf("failed to get free port: %v", err))
	}
	return port
}

// GetFreePorts returns n free port numbers
func GetFreePorts(n int) ([]int, error) {
	ports := make([]int, n)
	for i := 0; i < n; i++ {
		port, err := GetFreePort()
		if err != nil {
			return nil, err
		}
		ports[i] = port
	}
	return ports, nil
}

// MustGetFreePorts is like GetFreePorts but panics on error
func MustGetFreePorts(n int) []int {
	ports, err := GetFreePorts(n)
	if err != nil {
		panic(fmt.Sprintf("failed to get %d free ports: %v", n, err))
	}
	return ports
}

// PortsToStringMap converts a slice of ports to a map[string]string for environment variables
func PortsToStringMap(portNames []string, ports []int) map[string]string {
	envMap := make(map[string]string)
	for i, name := range portNames {
		if i < len(ports) {
			envMap[name] = strconv.Itoa(ports[i])
		}
	}
	return envMap
}
