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
