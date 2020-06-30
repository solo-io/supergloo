# To test the major refactor:

Make sure kind is set up with istio, then:

```bash
# register cluster
go run cmd/cli/main.go cluster register \
    --cluster-name master-cluster \
    --cluster-domain-override=host.docker.internal

# run discovery
go run cmd/mesh-discovery/main.go 
```
