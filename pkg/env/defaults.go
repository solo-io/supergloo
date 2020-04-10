package env

import (
	"os"
	"strconv"
)

const (
	EnvPodNamespace       = "POD_NAMESPACE"
	defaultWriteNamespace = "service-mesh-hub"

	EnvGrpcPort     = "GRPC_PORT"
	defaultGrpcPort = 10101

	EnvHealthCheckPort     = "HEALTH_CHECK_PORT"
	defaultHealthCheckPort = 8081
)

func GetWriteNamespace() string {
	writeNamespace := os.Getenv(EnvPodNamespace)
	if writeNamespace == "" {
		writeNamespace = defaultWriteNamespace
	}
	return writeNamespace
}

func GetGrpcPort() int {
	stringPort := os.Getenv(EnvGrpcPort)
	if stringPort == "" {
		stringPort = strconv.Itoa(defaultGrpcPort)
	}
	port, err := strconv.Atoi(stringPort)
	if err != nil {
		return defaultGrpcPort
	}
	return port
}


func GetHealthCheckPort() int {
	stringPort := os.Getenv(EnvHealthCheckPort)
	if stringPort == "" {
		stringPort = strconv.Itoa(defaultHealthCheckPort)
	}
	port, err := strconv.Atoi(stringPort)
	if err != nil {
		return defaultHealthCheckPort
	}
	return port
}
