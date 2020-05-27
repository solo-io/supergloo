package metadata

func BuildLocalFQDN(meshServiceName string) string {
	return meshServiceName + ".default.svc.cluster.local"
}
